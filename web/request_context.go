package web

import (
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/tools"
	"net/http"
	"reflect"
	"strings"
	"sync"
)

const (
	// Get请求参数：数组类型最大支持元素个数
	MAX_PARAM_LIST_LEN = 200

	// 请求方式
	HTTP_GET    = "GET"
	HTTP_POST   = "POST"
	HTTP_DELETE = "DELETE"
	HTTP_PUT    = "PUT"
	HTTP_PATCH  = "PATCH"

	// 返回结果支持的格式 format=xml、format=json，默认为json
	FORMAT_JSON = "json"
	FORMAT_XML  = "xml"

	// 普通get post请求，必须在url上的参数，format和version都有默认值
	REQ_ACTION      = "action"
	REQ_VERSION     = "version"
	DEFAULT_VERSION = "v1"
)

var apiErrorType = reflect.TypeOf(new(e.ApiError)).Elem()
var superErrorType = reflect.TypeOf(new(error)).Elem()
var timeType = reflect.TypeOf(new(basic.TimeStandard)).Elem()
var dayType = reflect.TypeOf(new(basic.Day)).Elem()

// 请求上下文，在当前请求中进行传递
type ReqContext struct {
	requestId    string                 //当前请求的唯一ID
	request      interface{}            //对应Controller入参
	response     interface{}            //对应Controller出参
	superRequest interface{}            //对应Controller入参中的通用参数
	model        string                 //对应Controller所在模块
	action       string                 //对应Controller名称
	version      string                 //对应Controller版本
	params       map[string][]string    //对应Get请求Url路径上的参数
	attr         map[string]interface{} //可以设置一些自定义缓存属性，在当前请求中一直有效
	method       string                 //对应http请求的类型：GET、POST
	status       int                    //http响应码
	r            *http.Request          //http请求
	format       string                 //格式化响应Response，支持json、xml
	mu           sync.RWMutex
}

func (ctx *ReqContext) SetRequestId(requestId string) {
	ctx.requestId = requestId
}
func (ctx *ReqContext) SetFormat(format string) {
	ctx.format = format
}
func (ctx *ReqContext) GetRequestId() string {
	return ctx.requestId
}
func (ctx *ReqContext) GetRequest() interface{} {
	return ctx.request
}
func (ctx *ReqContext) GetResponse() interface{} {
	return ctx.response
}
func (ctx *ReqContext) GetAttribute(key string) interface{} {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.attr[key]
}
func (ctx *ReqContext) SetAttribute(key string, val interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.attr[key] = val
}
func (ctx *ReqContext) GetHeader(key string) string {
	return ctx.r.Header.Get(key)
}
func (ctx *ReqContext) GetHttpRequest() *http.Request {
	return ctx.r
}

// 上下文中保存：Controller入参
func (ctx *ReqContext) SetRequest(model interface{}) {
	ctx.request = model
}

// 上下文中保存：Controller出参
func (ctx *ReqContext) SetResponse(model interface{}) {
	if model == nil {
		return
	}
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr && val.IsNil() {
		return
	}

	ctx.response = model
}

func (ctx *ReqContext) GetRemoteIp() string {
	return ctx.r.RemoteAddr
}
func (ctx *ReqContext) GetModel() string {
	return ctx.model
}
func (ctx *ReqContext) GetAction() string {
	return ctx.action
}
func (ctx *ReqContext) GetMethod() string {
	return ctx.method
}
func (ctx *ReqContext) GetSuperRequest() interface{} {
	return ctx.superRequest
}

// 获得方法名称
func (ctx *ReqContext) getActionName() string {
	return ctx.model + "." + ctx.action
}

// 上下文中获得请求参数值，只有Get方式请求时可用
func (ctx *ReqContext) GetParamValue(key string) string {

	// 获得数组中第一个值
	firstVal := func(vals []string) string {
		if vals == nil || len(vals) < 1 {
			return ""
		}
		return vals[0]
	}

	//遍历获得Value
	for k, v := range ctx.params {
		if strings.EqualFold(key, k) {
			return firstVal(v)
		}
	}

	return ""
}

// 上下文中保存：Controller入参中的通用参数
func (ctx *ReqContext) setSuperRequest(model interface{}) {
	ctx.superRequest = model
}

// 普通get或post请求，解析url获取model、action、version
// 创建Context
func (ctx *ReqContext) createContext(r *http.Request, conf *WebConf, useModel bool) error {

	// 得到请求路径
	// 如果设置的controller的UseModel为true， 那么请求为：/rootPath/model?Action=SomeMethod，计算出的 path=/rootPath/model
	// 如果设置的controller的UseModel为false，那么请求为：/rootPath?Action=SomeMethod，计算出的 path=/rootPath
	path := r.URL.Path

	// rootPath为可选的，如果配置文件中配置了，则必须指定，否则可忽略
	if conf != nil && conf.RootPath != "" {
		rootPath := "/" + conf.RootPath
		if !tools.MatchReg(path, rootPath+`(?:$|[?]|/)`) {
			return e.NewApiError(e.NOT_FOUND, "Missing RootPath in request url.", nil)
		} else {
			path = strings.Replace(path, rootPath, "", 1) // 截掉rootPath
		}
	}

	// 剩下的部分，去掉/之后，就是model部分了，但是如果controller的UseModel为false，这部分就是空的，需要去controller里面取真实的model
	// 不用model时此处暂时设置为空，待定位到具体controller时，再设置model
	path = strings.Replace(path, "/", "", 1)
	if useModel && strings.Index(path, "/") > -1 {
		return e.NewApiError(e.NOT_FOUND, "Invalid http request. Uri parsing failed.", nil)
	}

	// 请求参数列表
	firstVal := func(vals []string) string {
		if vals == nil || len(vals) != 1 {
			return ""
		}
		return vals[0]
	}

	// action
	action := firstVal(r.URL.Query()[REQ_ACTION])
	if action == "" {
		action = firstVal(r.URL.Query()[strings.Title(REQ_ACTION)])
	}
	if action == "" {
		return e.NewApiError(e.NOT_FOUND, "Invalid action name.", nil)
	}

	// version
	version := firstVal(r.URL.Query()[REQ_VERSION])
	if version == "" {
		version = firstVal(r.URL.Query()[strings.Title(REQ_VERSION)])
	}
	if version == "" {
		version = DEFAULT_VERSION
	}

	// Context
	ctx.model = path
	ctx.action = strings.Title(action)
	ctx.version = version

	return nil
}
