package restful

import (
	"regexp"
)

// restful协议通用属性封装

// 自定义类型
type protocol string
type service string

const (
	DELETE protocol = "DELETE"
	PUT    protocol = "PUT"
	GET    protocol = "GET"
	POST   protocol = "POST"
	PATCH  protocol = "PATCH"
)

const (
	FieldReg     = `\{([a-zA-Z0-9-_.\[\]]+)\}`     //匹配字段名称
	ValueReg     = `([\p{Han}a-zA-Z0-9-_.\[\]]+)`  //匹配字段值
	PathValueReg = `([\p{Han}a-zA-Z0-9-_.\[\]/]+)` //匹配带"/"符号的路径值
)

func NewRestful() *restfulMethod {
	return &restfulMethod{_intFields: make(map[string]interface{})}
}

// Restful协议描述
type restfulMethod struct {
	_serviceName service                // 接口所属产品线
	_method      protocol               // 支持的请求协议 GET、POST
	_versions    []string               // 支持的版本
	_action      string                 // 接口名称
	_uri         string                 // 接口匹配路径
	_uriReg      string                 // 接口匹配路径正则规则
	_fields      []string               // url路径上的参数占位，如/regions/{regionId}/instances/{instanceId}，得到的结果是 ["regionId","instanceId"]
	_intFields   map[string]interface{} // 数字类型字段
}

func (this *restfulMethod) ServiceName(serviceName service) *restfulMethod {
	this._serviceName = serviceName
	return this
}

func (this *restfulMethod) Method(method protocol) *restfulMethod {
	this._method = method
	return this
}

func (this *restfulMethod) Action(action string) *restfulMethod {
	this._action = action
	return this
}

func (this *restfulMethod) IntFields(fields ...string) *restfulMethod {
	for _, f := range fields {
		this._intFields[f] = struct{}{}
	}
	return this
}

// 设置uri
// 取出urk中的替换字段，同时设置uri正则路径
func (this *restfulMethod) Uri(uri string, pathFields ...string) *restfulMethod {

	// 从路径中取出所有变量字段
	this._uri = uri
	pattern := regexp.MustCompile(FieldReg)
	parts := pattern.FindAllStringSubmatch(uri, -1)
	this._fields = make([]string, 0)
	for _, v := range parts {
		if len(v) > 1 {
			this._fields = append(this._fields, v[1])
		}
	}
	// 将路径中的所有变量字段，替换为匹配Value的正则表达式
	this._uriReg = pattern.ReplaceAllStringFunc(uri, func(field string) string {
		for _, pathField := range pathFields {
			if pathField == field {
				return PathValueReg
			}
		}
		return ValueReg
	})
	this._uriReg += "(?:$|[?]|/$|/[?])"
	return this
}

// 设置支持的版本
func (this *restfulMethod) SupportVersion(versions ...string) *restfulMethod {
	if this._versions == nil {
		this._versions = make([]string, 0)
	}
	for _, v := range versions {
		this._versions = append(this._versions, v)
	}
	return this
}

// 判断是否支持此版本
func (this *restfulMethod) contains(version string) bool {
	for _, v := range this._versions {
		if v == version {
			return true
		}
	}
	return false
}
