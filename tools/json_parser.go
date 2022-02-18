package tools

import (
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"github.com/Zoxu0928/task-common/tools/json/jsonparser"
)

// 解析数组
func arrayHandler(value []byte, dataType jsonparser.ValueType, offset int, err error) {
	if dataType == jsonparser.Object {
		jsonparser.ObjectEach(value, objectHandler)
	}
}

// 解析对象
func objectHandler(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
	if key[0] >= 65 && key[0] <= 90 {
		key[0] = byte(key[0] + 32)
	}
	if dataType == jsonparser.Object {
		jsonparser.ObjectEach(value, objectHandler)
	} else if dataType == jsonparser.Array {
		jsonparser.ArrayEach(value, arrayHandler)
	}
	return nil
}

// 将对象解析为json，并且key的首字母转为小写
// 废弃，不要使用了
//func JsonUncamel(v interface{}) ([]byte, error) {
//    if v == nil {
//        return []byte{}, nil
//    }
//
//    // 禁止html转义
//    buffer := &bytes.Buffer{}
//    encoder := json.NewEncoder(buffer)
//    encoder.SetEscapeHTML(false)
//    if ee := encoder.Encode(v); ee != nil {
//        return nil, ee
//    }
//
//    // 去掉换行符
//    b := buffer.Bytes()[0 : buffer.Len()-1]
//    if b[0] == 123 {
//        if err := jsonparser.ObjectEach(b, objectHandler); err != nil {
//            return b, err
//        }
//    } else if b[0] == 91 {
//        if _, err := jsonparser.ArrayEach(b, arrayHandler); err != nil {
//            return b, err
//        }
//    }
//    return b, nil
//}

// 获取Json中Key的值
func FindJsonValue(json *string, keys ...string) string {
	if json == nil {
		return ""
	}
	val, _ := jsonparser.GetString([]byte(*json), keys...)
	return val
}

//=====================================================================================
// 将对象解析成Url参数
func ToUrl(v interface{}) (string, error) {
	if v == nil {
		return "", nil
	}
	b, _ := ffjson.Marshal(v)

	param_key := ""
	index_map := map[string]int{}
	url := ""
	split := ""

	var getIndex = func(key string) string {
		index, ok := index_map[key]
		if !ok {
			return ""
		}
		index = index + 1
		index_map[key] = index
		return "." + ToString(index)
	}

	var fixUrl = func(value []byte) {
		url = url + split + param_key + getIndex(param_key) + "=" + string(value)
		if split == "" {
			split = "&"
		}
	}

	// 解析数组和对象的Func
	var arrayFunc func(value []byte, dataType jsonparser.ValueType, offset int, err error)
	var objectFunc func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error
	// 处理数组
	arrayFunc = func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		if dataType == jsonparser.Object {
			param_key = param_key + getIndex(param_key)
			jsonparser.ObjectEach(value, objectFunc)
			param_key = Replace(param_key, "[.]{0,1}[0-9a-zA-Z]+$", "")
		} else {
			//fmt.Println("1", param_key + getIndex(param_key), string(value))
			fixUrl(value)
		}
	}
	// 处理Object
	objectFunc = func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		if key[0] >= 65 && key[0] <= 90 {
			key[0] = byte(key[0] + 32)
		}
		if param_key == "" {
			param_key = string(key)
		} else {
			param_key = param_key + "." + string(key)
		}
		switch dataType {
		case jsonparser.Object:
			jsonparser.ObjectEach(value, objectFunc)
		case jsonparser.Array:
			index_map[param_key] = 0
			jsonparser.ArrayEach(value, arrayFunc)
		default:
			//fmt.Println("2", param_key, string(value))
			fixUrl(value)
		}
		param_key = Replace(param_key, "[.]{0,1}[0-9a-zA-Z]+$", "")
		return nil
	}
	if b[0] == 123 {
		if err := jsonparser.ObjectEach(b, objectFunc); err != nil {
			return "", err
		}
	} else if b[0] == 91 {
		if _, err := jsonparser.ArrayEach(b, arrayFunc); err != nil {
			return "", err
		}
	}
	return url, nil
}
