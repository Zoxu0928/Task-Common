package web

import (
	"context"
	"fmt"
	"github.com/Zoxu0928/task-common/api/restful"
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/controller"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/interceptor"
	"github.com/Zoxu0928/task-common/logger"
	"net/http"
	"reflect"
	"time"
)

// 用途：使用该Web框架，可以很容易的创建web服务，并接收所有请求
// 实例化示例：
// webServer := web.Load("web.yaml")
// webServer.StartUp()
// 以上启动了一个空的web服务，其中没有任何接口可以调用

// Web框架中的组件：
// 1. bean - 实体类
//      就是业务代码中的某某Service，加入到bean组件中的Service可实现自动依赖注入。
//      使用bean之后就不需要再到处New对象，到处传递对象
// 2. controller - 路由接口
//      就是业务代码中的某某Service中的首字母大写的Func
// 3. interceptor - 拦截器
// 4. validator - 参数验证

// Web框架对请求的处理逻辑大概如下：
// 1. 接收 Get请求、Post json请求、restful请求
// 2. 实例化RequestContext请求上下文，里面可以获取一些通用属性，可以设置一些缓存信息
// 3. 定位Controller，实例化Controller的入参
// 4. 将http请求的参数转换为Controller的入参，只能转换正常参数，Header中的无法处理
// 5. 执行拦截器链（BeforeFunc），在拦截器中可以处理：限流、日志、参数校验等逻辑
// 6. 执行Controller
// 7. 反向执行拦截器链（AfterFunc）
// 8. 输出响应

// 如果没有配置参数，则使用默认配置参数启动web服务
// 默认端口 8080
// 默认读超时 30秒
// 默认写超时 30秒

// 默认配置
const (
	DEFAULT_HTTP_PORT          = "8080"
	DEFAULT_HTTP_READ_TIMEOUT  = 30 * time.Second
	DEFAULT_HTTP_WRITE_TIMEOUT = 30 * time.Second
)

// 可以向webServer中注册一些函数，以便在请求执行之前，或向客户端响应之前做一些自定义的处理

// Web框架收到http请求后，会第一时间进行回调。如果回调返回true，Web框架会结束请求。
type beforeDispatchCallbackFunc func(w http.ResponseWriter, r *http.Request) bool

// Web框架处理完http参数后，会回调一次，在回调中可以处理一些非常规的参数，比如Header中的自定义参数。
type handlerRequestFunc func(ctx *ReqContext, restAction *restful.ActionMeta, w http.ResponseWriter, r *http.Request) e.ApiError

// Web框架在输出Response之前，会回调一次，在回调中可以实现一些自定义逻辑
type handlerResponseFunc func(ctx *ReqContext, response interface{}) interface{}

// Web框架在输出ErrorResponse之前，会回调一次，在回调中可以实现一些自定义逻辑
type handlerErrorFunc func(ctx *ReqContext, err error) interface{}

// 输出响应之前回调
type beforeResponseFunc func(ctx *ReqContext)

// Web框架
type webServer struct {

	// web server 的配置信息
	conf *WebConf

	// web server
	server *http.Server

	// 可选：
	// 功能：在 http server 收知请求后，首先会偿试执行 beforeDispatch 函数。根据 beforeDispatch 函数的返回值，再决定请求继续执行或直接返回
	// Web框架在执行 dispatchHandler 将请求下发之前，会预先执行一次回调处理，处理逻辑自定义。如果不需要预处理，则不需要关心此参数
	// 如果返回true，流程会直接结束不会继续下发
	// 如果返回false，则下一步会调用 dispatchHandler 将请求下发
	// 应用场景：
	// 1. 自定义限流逻辑
	// 2. 非接口调用的请求。比如静态页面等，就不需要下发请求给接口
	// 3. 特殊场景的http请求逻辑，想要自己处理http请求
	beforeDispatch beforeDispatchCallbackFunc

	// 可选：
	// Web框架组装完成request参数之后，调用接口之前，会回调一次检查参数，检查逻辑需要用户自行实现。
	// 此处可以处理一些非常规参数的处理
	handlerRequest handlerRequestFunc

	// 可选：
	// 生成成功响应的数据，如不设置程序会按默认处理
	// Web框架响应返回客户端之前，会回调一次此接口，可以做一些自定义逻辑，比如改变响应的结构
	handlerResponse handlerResponseFunc

	// 可选：
	// 生成失败响应的数据，如不设置程序会按默认处理
	// Web框架响应返回客户端之前，会回调一次此接口，可以做一些自定义逻辑，比如改变响应的结构
	handlerError handlerErrorFunc

	// 响应输出之前，回调一次
	beforeResponse beforeResponseFunc

	// 我们需要把通用的参数抽到单独的struct中，由每个接口不同的request去继承通用的struct
	// 比如：
	// type SuperRequest struct { //通用参数为用户ID
	//     UserId
	// }
	// type SendMailRequest struct {
	//     SuperRequest
	//     ToUser string
	//     Mail   string
	// }
	// 上面的SuperRequest即通用的匿名struct，可选用以下两种方式设置：
	// 1. webServer.SetSuperRequestType("SuperRequest")
	// 2. webServer.SetSuperRequest(new(api.SuperRequest))
	// Web框架并不会使用SuperRequest，增加此功能的意义在于，在回调中、或者在拦截器中，使用者可以很容易的从RequestContext中获得这个通用参数
	// 大多数情况下，我们都会使用通用参数进行一些业务校验、或其它通用的逻辑处理，如果没有此通用参数的功能，做起来会极为不方便

	// 匿名struct名称
	superRequestType string

	// 匿名struct类型，名称与类型二选一即可
	superRequest reflect.Type

	// 可选：
	// 拦截器链表
	// 如果没有设置，Web框架会执行全局默认的拦截器链表
	// 如果设置了，Web框架只会执行绑定在自己端口下的拦截器链表
	bindInterceptor interceptor.Interceptor

	// 可选：
	// Controller路由
	// 如果没有设置，Web框架会使用默认的全局路由定位Controller
	// 如果设置了，Web框架只会在绑定在自己端口下的路由中查找Controller
	bindController controller.Controller
}

// 加载配置文件并返回web实例
//func Load(cfgPath string) *webServer {
//	web := &webServer{}
//	web.conf = web.loadWebConfig(cfgPath)
//	return web
//}

func Load(conf *WebConf) *webServer {
	parseCfg(conf)
	return &webServer{conf: conf}
}

// web配置
type WebConf struct {
	Name         string         `yaml:"http.name" toml:"name"` // 可空
	Port         string         `yaml:"http.port" toml:"port"`
	ReadTimeout  basic.Duration `yaml:"http.readTimeout" toml:"read_timeout"`
	WriteTimeout basic.Duration `yaml:"http.writeTimeout" toml:"write_timeout"`
	RootPath     string         `yaml:"http.rootPath" toml:"root_path"` // 可空
	Models       []string       `yaml:"http.models" toml:"models"`      // 可空，如果不为空，说明此httpServer只能访问配置的模块
}

func (this *webServer) GetConf() *WebConf {
	return this.conf
}
func (this *webServer) SetPort(port string) *webServer {
	this.conf.Port = port
	return this
}
func (this *webServer) SetReadTimeout(v basic.Duration) *webServer {
	this.conf.ReadTimeout = v
	return this
}
func (this *webServer) SetWriteTimeout(v basic.Duration) *webServer {
	this.conf.WriteTimeout = v
	return this
}
func (this *webServer) SetSuperRequestType(superRequestType string) *webServer {
	this.superRequestType = superRequestType
	return this
}
func (this *webServer) SetSuperRequest(v interface{}) *webServer {
	this.superRequest = reflect.TypeOf(v).Elem()
	return this
}
func (this *webServer) SetBeforeDispatch(f beforeDispatchCallbackFunc) *webServer {
	this.beforeDispatch = f
	return this
}
func (this *webServer) SetCheckRequest(f handlerRequestFunc) *webServer {
	this.handlerRequest = f
	return this
}
func (this *webServer) SetHandlerResponse(f handlerResponseFunc) *webServer {
	this.handlerResponse = f
	return this
}
func (this *webServer) SetHandlerError(f handlerErrorFunc) *webServer {
	this.handlerError = f
	return this
}
func (this *webServer) SetBeforeResponse(f beforeResponseFunc) *webServer {
	this.beforeResponse = f
	return this
}
func (this *webServer) BindInterceptor(i interceptor.Interceptor) *webServer {
	this.bindInterceptor = i
	return this
}
func (this *webServer) BindController(c controller.Controller) *webServer {
	this.bindController = c
	return this
}

//================================================================================
// 校验或初始化Http配置参数
//================================================================================
func parseCfg(httpConf *WebConf) {
	if httpConf.Port == "" {
		httpConf.Port = DEFAULT_HTTP_PORT
	}
	if httpConf.ReadTimeout.Duration <= 0 {
		httpConf.ReadTimeout.Duration = DEFAULT_HTTP_READ_TIMEOUT
	}
	if httpConf.WriteTimeout.Duration <= 0 {
		httpConf.WriteTimeout.Duration = DEFAULT_HTTP_WRITE_TIMEOUT
	}
}

// 核心Handler，所有进入的请求都会经过此handler
type commonHandler struct {
	web *webServer
}

// 实现ServeHTTP，代表commonHandler实现了Handler
func (h *commonHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// 下发请求之前，预处理一次
	// 如果预处理的结果为true，说明上方业务层自己处理了该请求，该框架不再处理该请求，流程直接结束，不再处理
	if h.web.beforeDispatch != nil && h.web.beforeDispatch(w, r) {
		return
	}

	// Api接口请求：下发请求
	h.dispatchHandler(w, r)
}

//================================================================================
// 创建Web服务，开启监听
//================================================================================
func (web *webServer) StartUp() {

	// server
	web.server = &http.Server{
		Addr:           ":" + web.conf.Port,
		ReadTimeout:    web.conf.ReadTimeout.Duration,
		WriteTimeout:   web.conf.WriteTimeout.Duration,
		Handler:        &commonHandler{web: web},
		MaxHeaderBytes: 1 << 20,
	}

	logger.Info(fmt.Sprintf(""+
		"\n*************************************************************"+
		"\n*********** web service [name:%s] [root-path:%s] start listening on %s *************"+
		"\n*************************************************************", web.conf.Name, web.conf.RootPath, web.conf.Port))
	if len(web.conf.Models) > 0 {
		logger.Info("Port %s Support models %s", web.conf.Port, web.conf.Models)
	}

	// 开始监听
	go func() {
		err := web.server.ListenAndServe()
		if err != nil {
			if err.Error() != "http: Server closed" {
				panic(err)
			}
		}
	}()
}

func (web *webServer) Close() {
	logger.Info("[Application close] Web server [:%s] shutting down.", web.conf.Port)
	if web.server != nil {
		ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
		err := web.server.Shutdown(ctx)
		if err != nil {
			logger.Info("Web server stop error: %v", err)
		} else {
			logger.Info("Web server gracefully stopped")
		}
	}
}
