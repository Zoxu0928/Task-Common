package restful

import (
	"github.com/Zoxu0928/task-common/tools"
	"net/http"
	"regexp"
	"strings"
)

// 所有接口列表
var apis = make([]*restfulMethod, 0)

// 增加注册一个接口
func RegisterApi(method *restfulMethod) {
	apis = append(apis, method)
}

// restful url分解后的模型
type action struct {
	method      string // GET、POST
	serviceName string // 模块
	version     string // v1
	name        string // 接口名称
}

// restful 具体请求的内容，包括参数名称和值
type ActionMeta struct {
	action
	params map[string]interface{} // url上的其它参数的key-value列表
}

func (this *ActionMeta) GetMethod() string {
	return this.method
}
func (this *ActionMeta) GetServiceName() string {
	return this.serviceName
}
func (this *ActionMeta) GetVersion() string {
	return this.version
}
func (this *ActionMeta) GetName() string {
	return this.name
}
func (this *ActionMeta) GetParams() map[string]interface{} {
	return this.params
}
func (this *ActionMeta) PutParam(key string, value interface{}) {
	if this.params == nil {
		this.params = make(map[string]interface{})
	}
	this.params[key] = value
}

// 根据url匹配api接口
func Match(method, path string) *ActionMeta {

	// 拆分为版本和请求路径
	request_version := getVersion(path)
	request_path := getPath(path)

	// 遍历所有api
	for _, _api := range apis {

		// 如果api支持此方式
		if string(_api._method) == method {

			// 匹配请求路径
			pattern := regexp.MustCompile(_api._uriReg)
			parts := pattern.FindStringSubmatch(request_path)

			// 如果匹配成功
			if len(parts) > 0 {

				// 如果不支持此版本
				if !_api.contains(request_version) {
					continue
				}
				a := &ActionMeta{}
				a.method = string(_api._method)
				a.serviceName = string(_api._serviceName)
				a.version = request_version
				a.name = _api._action
				a.params = make(map[string]interface{})
				// 封装url上需要替换的变量
				for i := 1; i < len(parts); i++ {
					fieldName := getvalue(_api._fields, i-1)
					if fieldName != "" {
						if fieldValue := getvalue(parts, i); fieldValue != "" {
							if _, ok := _api._intFields[fieldName]; ok {
								intVal, _ := tools.ToInt(fieldValue)
								a.params[fieldName] = intVal
							} else {
								a.params[fieldName] = fieldValue
							}
						}
					}
				}
				return a
			}
		}
	}
	return nil
}

// 根据下标获取数组值
func getvalue(list []string, index int) string {
	if len(list) > index {
		return list[index]
	}
	return ""
}

// 拆分path获得version
func getVersion(path string) string {
	if len(path) < 1 {
		return ""
	}
	if strings.Index(path, "/") == 0 {
		path = path[1:]
	}
	return path[0:strings.Index(path, "/")]
}

// 拆分path获得version以后的路径
func getPath(path string) string {
	if len(path) < 1 {
		return ""
	}
	if strings.Index(path, "/") == 0 {
		path = path[1:]
	}
	return path[strings.Index(path, "/"):]
}

// 检查请求是否是restful方式
func IsRestfulRequest(request *http.Request) bool {
	pattern := regexp.MustCompile(`^/v[0-9]+/`)
	return pattern.Match([]byte(request.RequestURI))
}
