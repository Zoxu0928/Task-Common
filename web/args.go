package web

import (
	"encoding/json"
	"fmt"
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/tools"
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"github.com/Zoxu0928/task-common/tools/json/json-iterator/go"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

// 转换http请求的参数，封装为目标请求参数
func (h *commonHandler) createArg(r *http.Request, arg reflect.Value, ctx *ReqContext) error {

	// GET参数
	if r.Method == HTTP_GET || r.Method == HTTP_DELETE {

		// 是否指针的处理
		stru := arg
		if stru.Kind() == reflect.Ptr {
			stru = stru.Elem()
		}

		// GET方式处理 url parameters
		if _, err := h.createGetArg(stru, ctx, "", ""); err != nil {
			return err
		}

		// POST参数
	} else if r.Method == HTTP_POST || r.Method == HTTP_PUT || r.Method == HTTP_PATCH {

		// 首先处理body流，json body
		if err := h.createPostJsonArg(r, arg, ctx); err != nil {
			return err
		}

		// 如果post请求的url上有parameters，还需要处理一次url上的参数
		if len(ctx.params) > 0 {
			stru := arg
			if stru.Kind() == reflect.Ptr {
				stru = stru.Elem()
			}
			if _, err := h.createGetArg(stru, ctx, "", ""); err != nil {
				return err
			}
		}
	} else {
		return e.NewApiError(e.UNAVAILABLE, fmt.Sprintf("Http method %s is not supported.", r.Method), nil)
	}
	ctx.SetRequest(arg.Interface())
	return nil
}

// 封装get请求参数
func (h *commonHandler) createGetArg(arg reflect.Value, ctx *ReqContext, pName string, superName string) (bool, error) {

	obj_changed := false

	// 遍历字段
	for i := 0; i < arg.NumField(); i++ {

		// 获得字段
		f := arg.Field(i)

		// 忽略函数类型的字段
		if f.Kind() == reflect.Func {
			continue
		}

		// 如果字段是匿名下来的，计算匿名名字，anonymousName代表的就是匿名的名称部分
		afType := arg.Type().Field(i)
		anonymousName := superName
		if afType.Anonymous {
			anonymousName = anonymousName + afType.Name + "."
		}

		// 如果字段类型是指针，取其实际的值，实际的值肯定是nil，因为还没有初始化
		real_field := f
		if real_field.Kind() == reflect.Ptr {
			real_field = real_field.Elem()
		}

		// 如果字段是指针类型，那么值肯定是nil的
		// 如果字段是普通类型的指针，那么忽略
		// 如果字段是struct类型的指针，则先暂时初始化一个struct，如果后面struct中没有任何字段被赋值，则丢弃，否则将这个struct赋给其所属字段
		// 意思就是指针类型的字段，如果指定了值则不为空，否则这个字段就是nil的
		var new_obj_flag bool = false
		var new_obj_val *reflect.Value
		if f.Kind() == reflect.Ptr && real_field.IsValid() == false {
			obj := reflect.New(f.Type().Elem())
			if obj.Elem().Kind() == reflect.Struct {
				real_field = obj.Elem()
				new_obj_flag = true
				new_obj_val = &obj
			}
		}

		// 如果是通用request，缓存super对象到context
		if real_field.Kind() == reflect.Struct {
			if real_field.Type().Name() == h.web.superRequestType || (h.web.superRequest != nil && real_field.Type().AssignableTo(h.web.superRequest)) {
				ctx.setSuperRequest(real_field.Addr().Interface())
			}
		}

		// 如果是内部对象，递归
		if real_field.Kind() == reflect.Struct {

			// 如果该字段为时间类型，直接赋值
			if real_field.Type().AssignableTo(timeType) {
				if _changed, _err := h.setTimeStandard(afType.Name, pName, anonymousName, ctx, &f); _err != nil {
					return obj_changed, _err
				} else {
					obj_changed = _changed
				}
				continue
			}

			// 如果该字段为日期类型，直接赋值
			if real_field.Type().AssignableTo(dayType) {
				if _changed, _err := h.setDay(afType.Name, pName, anonymousName, ctx, &f); _err != nil {
					return obj_changed, _err
				} else {
					obj_changed = _changed
				}
				continue
			}

			// 其它结构体，递归
			if changed, err := h.createGetArg(real_field, ctx, h.getFieldName(arg.Type().Field(i).Name, pName), anonymousName); err != nil {
				return changed, err
			} else if changed && new_obj_flag {
				// 如果是新的obj并且有改变，为指针变量赋值，否则指针变量为nil
				f.Set(*new_obj_val)
				obj_changed = true
				continue
			}

			// 如果是内部对象数组，递归
		} else if real_field.Kind() == reflect.Slice && real_field.Type().Elem().Kind() == reflect.Struct {

			// 生成一个空对象
			_eData := fmt.Sprintf("%s", reflect.New(real_field.Type().Elem()).Elem())
			for p := 0; p < MAX_PARAM_LIST_LEN; p++ {

				// 设置值
				item := reflect.New(real_field.Type().Elem()).Elem()
				if _, err := h.createGetArg(item, ctx, fmt.Sprintf("%s.%d", h.getFieldName(arg.Type().Field(i).Name, pName), p+1), anonymousName); err != nil {
					return false, err
				}

				// 如果和空对象相等说明到了数组结尾了
				_cData := fmt.Sprintf("%s", item)
				if _eData == _cData {
					break
				}

				// 加入数组
				real_field.Set(reflect.Append(real_field, item))
			}

			// 如果是内部对象指针数组，递归
		} else if real_field.Kind() == reflect.Slice && real_field.Type().Elem().Kind() == reflect.Ptr && real_field.Type().Elem().Elem().Kind() == reflect.Struct {

			// 生成一个空对象
			_eData := fmt.Sprintf("%s", reflect.New(real_field.Type().Elem().Elem()).Elem())
			for p := 0; p < MAX_PARAM_LIST_LEN; p++ {

				// 设置值
				item := reflect.New(real_field.Type().Elem().Elem()).Elem()
				if _, err := h.createGetArg(item, ctx, fmt.Sprintf("%s.%d", h.getFieldName(arg.Type().Field(i).Name, pName), p+1), anonymousName); err != nil {
					return false, err
				}

				// 如果和空对象相等说明到了数组结尾了
				_cData := fmt.Sprintf("%s", item)
				if _eData == _cData {
					break
				}

				// 加入数组
				real_field.Set(reflect.Append(real_field, item.Addr()))
			}

			// 普通字段
		} else {

			// 获得字段名称
			fieldName := h.getFieldName(afType.Name, pName)

			// 如果是字段是super中匿名继承下来的，去掉匿名部分
			// 比如Field.Base.Id，前面的Base是匿名类的名称，需要去掉，anonymousName代表的就是匿名部分
			if anonymousName != "" {
				fieldName = strings.Replace(fieldName, anonymousName, "", 1)
			}

			// 根据类型设置Value
			switch f.Type().String() {
			case "string", "*string":
				val := strings.TrimSpace(ctx.GetParamValue(fieldName))
				if val == "" {
					continue
				}
				if f.Kind() == reflect.Ptr {
					f.Set(reflect.ValueOf(&val))
				} else {
					f.SetString(val)
				}
				obj_changed = true
				break
			case "[]string":
				for i := 0; i < MAX_PARAM_LIST_LEN; i++ {
					val := strings.TrimSpace(ctx.GetParamValue(fmt.Sprintf("%s.%d", fieldName, i+1)))
					if val == "" {
						break
					}
					f.Set(reflect.Append(f, reflect.ValueOf(val)))
				}
				obj_changed = true
				break
			case "int", "*int", "int32", "*int32", "int64", "*int64":
				val := ctx.GetParamValue(fieldName)
				if val == "" {
					continue
				}
				if toVal, err := tools.ToInt(val); err != nil {
					return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Malformed %s %s", fieldName, val), nil)
				} else {
					if f.Kind() == reflect.Ptr {
						f.Set(reflect.ValueOf(&toVal))
					} else {
						f.SetInt(int64(toVal))
					}
				}
				obj_changed = true
				break
			case "[]int", "[]int32", "[]int64":
				for i := 0; i < MAX_PARAM_LIST_LEN; i++ {
					val := ctx.GetParamValue(fmt.Sprintf("%s.%d", fieldName, i+1))
					if val == "" {
						break
					}
					if toVal, err := tools.ToInt(val); err != nil {
						return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Malformed %s %s", fieldName, val), nil)
					} else {
						f.Set(reflect.Append(f, reflect.ValueOf(toVal)))
					}
				}
				obj_changed = true
				break
			case "bool", "*bool":
				val := ctx.GetParamValue(fieldName)
				if val == "" {
					continue
				}
				if toVal, err := tools.ToBool(val); err != nil {
					return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Malformed %s %s", fieldName, val), nil)
				} else {
					if f.Kind() == reflect.Ptr {
						f.Set(reflect.ValueOf(&toVal))
					} else {
						f.SetBool(toVal)
					}
				}
				obj_changed = true
				break
			case "[]bool":
				for i := 0; i < MAX_PARAM_LIST_LEN; i++ {
					val := ctx.GetParamValue(fmt.Sprintf("%s.%d", fieldName, i+1))
					if val == "" {
						break
					}
					if toVal, err := tools.ToBool(val); err != nil {
						return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Malformed %s %s", fieldName, val), nil)
					} else {
						f.Set(reflect.Append(f, reflect.ValueOf(toVal)))
					}
				}
				obj_changed = true
				break
			case "interface {}": // 这种类型的Get参数，全部转换为string，或者[]string

				// 先偿试按 string 处理
				if val := ctx.GetParamValue(fieldName); val != "" {
					f.Set(reflect.ValueOf(val))
					obj_changed = true
					break
				} else {
					// 偿试按[]string处理
					var _slice reflect.Value
					for i := 0; i < MAX_PARAM_LIST_LEN; i++ {
						val := ctx.GetParamValue(fmt.Sprintf("%s.%d", fieldName, i+1))
						if val == "" {
							break
						}
						if i == 0 && val != "" {
							_slice = reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0)
						}
						_slice = reflect.Append(_slice, reflect.ValueOf(val))
						obj_changed = true
					}
					if _slice.IsValid() && obj_changed {
						f.Set(_slice)
					}
					break
				}
			case "map[string]interface {}":
			default:
				return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Parse data error. Field %s data type is not supported", fieldName), nil)
			}
		}
	}
	return obj_changed, nil
}

// 封装post请求参数
func (h *commonHandler) createPostJsonArg(r *http.Request, arg reflect.Value, ctx *ReqContext) error {

	// 文件上传
	if strings.Index(r.Header.Get("Content-Type"), "multipart/form-data") > -1 {
		return nil
	}
	var body []byte
	if r.ContentLength > 0 {
		body, _ = ioutil.ReadAll(r.Body)
	} else {
		body = []byte("{}")
	}
	if err := ffjson.Unmarshal(body, arg.Interface()); err != nil {
		errMsg := ""
		if umErr, ok := err.(*json.UnmarshalTypeError); ok {
			errMsg = "Malformed parameter " + umErr.Field
		} else if umErr2, ok2 := err.(*jsoniter.UnmarshalTypeError); ok2 {
			errMsg = "Malformed parameter " + umErr2.Field
		}
		return e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Invalid request body. %s", errMsg), err)
	}
	h.injectSuperRequest(arg, ctx)
	return nil
}

func (h *commonHandler) injectSuperRequest(arg reflect.Value, ctx *ReqContext) bool {
	stru := arg
	if stru.Kind() == reflect.Ptr {
		stru = stru.Elem()
	}
	for i := 0; i < stru.NumField(); i++ {
		f := stru.Field(i)
		if f.Kind() == reflect.Struct {
			if f.Type().Name() == h.web.superRequestType || (h.web.superRequest != nil && f.Type().AssignableTo(h.web.superRequest)) {
				ctx.setSuperRequest(f.Addr().Interface())
				return true
			} else {
				if h.injectSuperRequest(f, ctx) {
					return true
				}
			}
		}
	}
	return false
}

// 时间类型字段设置
func (h *commonHandler) setTimeStandard(typeName, packageName, anonymousName string, ctx *ReqContext, field *reflect.Value) (bool, error) {

	obj_changed := false

	// 获得字段名称
	fieldName := h.getFieldName(typeName, packageName)

	// 如果是字段是super中匿名继承下来的，去掉匿名部分
	// 比如Field.Base.Id，前面的Base是匿名类的名称，需要去掉，anonymousName代表的就是匿名部分
	if anonymousName != "" {
		fieldName = strings.Replace(fieldName, anonymousName, "", 1)
	}

	val := ctx.GetParamValue(fieldName)
	if val == "" {
		return obj_changed, nil
	}
	time, timeErr := basic.NewTimeStandard(val)
	if timeErr != nil {
		return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Invalid %s %s", fieldName, val), nil)
	}
	if field.Kind() == reflect.Ptr {
		field.Set(reflect.ValueOf(&time))
	} else {
		field.Set(reflect.ValueOf(time))
	}
	obj_changed = true
	return obj_changed, nil
}

// 日期类型字段设置
func (h *commonHandler) setDay(typeName, packageName, anonymousName string, ctx *ReqContext, field *reflect.Value) (bool, error) {

	obj_changed := false

	// 获得字段名称
	fieldName := h.getFieldName(typeName, packageName)

	// 如果是字段是super中匿名继承下来的，去掉匿名部分
	// 比如Field.Base.Id，前面的Base是匿名类的名称，需要去掉，anonymousName代表的就是匿名部分
	if anonymousName != "" {
		fieldName = strings.Replace(fieldName, anonymousName, "", 1)
	}

	val := ctx.GetParamValue(fieldName)
	if val == "" {
		return obj_changed, nil
	}
	time, timeErr := basic.NewDay(val)
	if timeErr != nil {
		return obj_changed, e.NewApiError(e.INVALID_ARGUMENT, fmt.Sprintf("Invalid %s %s", fieldName, val), nil)
	}
	if field.Kind() == reflect.Ptr {
		field.Set(reflect.ValueOf(&time))
	} else {
		field.Set(reflect.ValueOf(time))
	}
	obj_changed = true
	return obj_changed, nil
}

// 拼接父子级别名称
func (h *commonHandler) getFieldName(name string, pName string) string {
	fieldName := name
	if pName != "" {
		fieldName = pName + "." + fieldName
	}
	return fieldName
}
