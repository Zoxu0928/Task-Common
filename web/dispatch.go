package web

// 用途：参数封装与请求下发

import (
	"fmt"
	"github.com/Zoxu0928/task-common/api/restful"
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/controller"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/interceptor"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"
	"github.com/Zoxu0928/task-common/tools/json/ffjson"

	"net/http"
	"net/url"
	"reflect"
)

// 入口
// 请求分发
// 支持普通 GET、POST 请求
// 支持 RESTFUL 请求
func (h *commonHandler) dispatchHandler(w http.ResponseWriter, r *http.Request) {

	// ------------------------------------------------------------------------
	// 第一步：接收到请求后，实始化 Context，此Context在请求执行到Controller之前，会一直传递

	// 为当前请求初始化一个Request Context
	ctx := &ReqContext{
		attr:      map[string]interface{}{},
		requestId: tools.GetGuid(),
		method:    r.Method,
		params:    r.URL.Query(),
		status:    http.StatusOK,
		r:         r,
	}

	func() {
		defer e.OnError("decode url")
		_url, _ := url.PathUnescape(r.URL.String())
		logger.Debug("%s Request-Url [%s] %s", ctx.requestId, r.Method, _url)
	}()

	// 第一时间设置响应格式，此参数的正确使用方式是在url的参数传过来
	ctx.format = ctx.GetParamValue("format")

	// 拦截panic，不管发生任何问题，都需要输出响应
	defer h.onError(w, ctx)

	// 判断当前请求是否是Restful方式
	is_restful_request := restful.IsRestfulRequest(r)
	var restAction *restful.ActionMeta

	// 如果是restful方式请求
	if is_restful_request {

		// 查找支持此Restful的接口
		path, err := url.PathUnescape(r.RequestURI)
		if err != nil {
			h.handlerError(w, e.NewApiError(e.INTERNAL, "Internal server error", err), ctx, true)
		}
		restAction = restful.Match(r.Method, path)

		// 提取restful接口描述信息
		if restAction != nil {

			ctx.action = restAction.GetName()
			ctx.model = restAction.GetServiceName()
			ctx.version = restAction.GetVersion()

		} else {
			h.handlerError(w, e.NewApiError(e.NOT_FOUND, "No such api.", nil), ctx, true)
			return
		}

		// 如果不是restful请求
	} else {

		// 普通GET、POST
		err := ctx.createContext(r, h.web.conf, h.useModel())
		if err != nil {
			h.handlerError(w, err, ctx, true)
			return
		}
	}

	// ------------------------------------------------------------------------
	// 第二步：定位后端Controller

	// 如果当前服务绑定了Controller，则在当前绑定下寻找。如果没有绑定，去全局找
	var method *basic.Method
	if h.web.bindController != nil {
		method = h.web.bindController.GetController(ctx.getActionName(), ctx.version)
	} else {
		method = controller.GetDefaultController().GetController(ctx.getActionName(), ctx.version)
	}

	if method == nil {
		if is_restful_request {
			h.handlerError(w, e.NewApiError(e.NOT_FOUND, fmt.Sprintf("No such api. model=%s api=%s version=%s", ctx.GetModel(), ctx.GetAction(), ctx.version), nil), ctx, true)
		} else {
			h.handlerError(w, e.NewApiError(e.NOT_FOUND, fmt.Sprintf("No such api. model=%s action=%s version=%s", ctx.GetModel(), ctx.action, ctx.version), nil), ctx, true)
		}
		return
	}

	// 如果useModel为false，此刻上下文中的model肯定是空的，需要从method中获取
	if h.useModel() == false && ctx.model == "" {
		ctx.model = method.GetPkgName()
	}

	// ------------------------------------------------------------------------
	// 第三步：验证该httpServer的模块权限
	if h.hasPermission(ctx) == false {
		logger.Error("%s port:%s Access controller %s.%s denied", ctx.requestId, h.web.conf.Port, ctx.model, ctx.action)
		h.handlerError(w, e.NewApiError(e.PERMISSION_DENIED, "Access denied.", nil), ctx, true)
		return
	}

	// ------------------------------------------------------------------------
	// 第四步：为该Controller封装入参

	// 封装入参
	args := make([]reflect.Value, 1)
	if len(method.GetArgs()) > 0 {
		firstArg := method.GetArgs()[0]
		if firstArg.Kind() == reflect.Ptr {
			args[0] = reflect.New(firstArg.Elem()) // 参数是指针
		} else if firstArg.Kind() == reflect.Struct {
			args[0] = reflect.New(firstArg) // 参数不是指针
		}
		if err := h.createArg(r, args[0], ctx); err != nil {
			h.handlerError(w, err, ctx, true)
			return
		}
		//a := args[0].Interface().(*jcs_model.DescribeInstanceRequest)
		//fmt.Println(tools.ToJson(a))
	}

	// 如果是restful请求，一些参数在url和header中，需要补充到request参数里面
	if is_restful_request {
		params, _ := ffjson.Marshal(restAction.GetParams())
		if rErr := ffjson.Unmarshal(params, args[0].Interface()); rErr != nil {
			logger.Error("%s unmarshal to request failed. %s", ctx.requestId, rErr)
		}
	}

	// 参数校验回调，在正式调用接口之前，可以做一些特殊处理，比如填充一些参数
	if h.web.handlerRequest != nil {
		if hre := h.web.handlerRequest(ctx, restAction, w, r); hre != nil {
			h.handlerError(w, hre, ctx, true)
			return
		}
	}

	// 如果是restful请求，一些参数在url和header中，需要补充到request参数里面
	if is_restful_request {
		params, _ := ffjson.Marshal(restAction.GetParams())
		if rErr := ffjson.Unmarshal(params, args[0].Interface()); rErr != nil {
			logger.Error("%s unmarshal to request failed. %s", ctx.requestId, rErr)
		}
	}

	// ------------------------------------------------------------------------
	// 第五步：
	// 截止到此处，目标Controller以及目标Controller的参数都已经具备了，并且有权限访问该模块
	// 开始执行链条中的拦截器，并最终调用目标Controller，再反向执行链条中的拦截器

	// 最终调用目标函数
	doFinal := func() error {
		return h.callTarget(method, args, ctx)
	}

	// 如果当前服务绑定了拦截器，则只执行当前绑定的。如果没有绑定，执行全局的
	if h.web.bindInterceptor != nil {
		if err := h.web.bindInterceptor.DoIntercepptor(ctx, method, doFinal); err != nil {
			h.handlerError(w, err, ctx, false)
			return
		}
	} else {
		if err := interceptor.GetDefaultInterceptor().DoIntercepptor(ctx, method, doFinal); err != nil {
			h.handlerError(w, err, ctx, false)
			return
		}
	}

	// ------------------------------------------------------------------------
	// 第六步：响应

	h.writeResponse(w, ctx.response, ctx, true)
}

// 关联的controller是否使用了model，如果使用了，http请求路径中需要带这部分数据，如果没有使用，http请求路径中不需要带这部分数据
// 正常请求：http://api.jd.com/RootPath/Model?Action=FuncName
// 不带model请求：http://api.jd.com/RootPath?Action=FuncName
func (h *commonHandler) useModel() bool {
	if h.web.bindController != nil {
		return h.web.bindController.UseModel()
	} else {
		return controller.GetDefaultController().UseModel()
	}
}

// 校验此端口的http服务是否有权限访问某model
func (h *commonHandler) hasPermission(ctx *ReqContext) bool {
	if len(h.web.conf.Models) > 0 {
		return tools.ContainsString(h.web.conf.Models, ctx.GetModel())
	}
	return true
}

// 调用目标函数
func (h *commonHandler) callTarget(method *basic.Method, args []reflect.Value, ctx *ReqContext) error {
	outs := method.Invoke(args)
	outTypes := method.GetReturn()
	for i, tp := range outTypes {
		if tp.AssignableTo(apiErrorType) {
			err := outs[i].Interface()
			if err != nil {
				return err.(e.ApiError)
			}
		} else if tp.AssignableTo(superErrorType) {
			err := outs[i].Interface()
			if err != nil {
				return err.(error)
			}
		} else {
			ctx.SetResponse(outs[i].Interface())
		}
	}
	return nil
}
