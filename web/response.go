package web

import (
	"errors"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"
	"net/http"
	"runtime/debug"
	"strings"

	safejson "github.com/helloeave/json"
)

// 捕获未知错误
func (h *commonHandler) onError(w http.ResponseWriter, ctx *ReqContext) {
	if r := recover(); r != nil {
		logger.Error("%s - 发生未知错误：%s\n%s", ctx.requestId, r, string(debug.Stack()))
		var err error
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("内部错误")
		}
		ctx.status = e.INTERNAL.Code
		h.handlerError(w, err, ctx, false)
	}
}

// 处理错误
func (h *commonHandler) handlerError(w http.ResponseWriter, err error, ctx *ReqContext, log bool) {

	if log {
		logger.Error("%s %s", ctx.requestId, err)
	}

	// request id
	requestId := ""
	if ctx == nil || ctx.requestId == "" {
		requestId = tools.GetGuid()
		ctx.requestId = requestId
	} else {
		requestId = ctx.requestId
	}

	// 如果设置了自定义错误响应，则直接回调，直接返回自定义的数据结构
	if h.web.handlerError != nil {
		h.writeResponse(w, h.web.handlerError(ctx, err), ctx, false)
		return
	}

	// 错误结构
	errResponse := &errorResponse{
		RequestId: requestId,
	}

	// 错误处理
	switch err.(type) {

	// OpenAip错误
	case e.ApiError:
		myErr := err.(e.ApiError)
		errResponse.Error.Code = myErr.GetCode()
		errResponse.Error.Status = myErr.GetType()
		errResponse.Error.Message = myErr.GetMessage()
		errResponse.Error.Details = myErr.GetDetails()
		ctx.status = myErr.GetCode()

		// 含有交易错误Code，放在返回的Header中
		if myErr.GetOErrCode() != "" {
			w.Header().Set("OrderErrCode", myErr.GetOErrCode())
		}

		// 其它错误
	default:
		errResponse.Error.Code = e.INTERNAL.Code
		errResponse.Error.Status = e.INTERNAL.Type
		errResponse.Error.Message = "Internal server error."
		ctx.status = e.INTERNAL.Code
	}

	h.writeResponse(w, errResponse, ctx, false)
}

// 请求失败响应的封装
type errorResponse struct {
	RequestId string
	Error     struct {
		Code    int
		Status  string
		Message string
		Details []map[string]string `json:",omitempty"`
	}
}

// 请求成功响应的封装
type successResponse struct {
	RequestId string      `json:"requestId"`
	Result    interface{} `json:"result,omitempty"`
}

// 将最终结果返回客户端
func (h *commonHandler) writeResponse(w http.ResponseWriter, response interface{}, ctx *ReqContext, success bool) {

	// 如果是下载文件，直接返回
	switch w.Header().Get("Content-Type") {
	case "application/octet-stream":
		return
	}

	// 非下载类操作，处理响应内容
	defer e.OnError(ctx.requestId)

	var data interface{}
	if success {
		// 如果设置了自定义处理响应，则使用自定义逻辑获取响应的数据结构
		if h.web.handlerResponse != nil {
			data = h.web.handlerResponse(ctx, response)
		} else {
			data = &successResponse{
				RequestId: ctx.requestId,
				Result:    response,
			}
		}
	} else {
		data = response
	}

	// 判断解析格式
	format := FORMAT_JSON
	if ctx.format != "" {
		format = ctx.format
	}

	if h.web.beforeResponse != nil {
		h.web.beforeResponse(ctx)
	}

	// 解析
	var res []byte
	var err error = nil
	var contentType string
	if strings.EqualFold(FORMAT_XML, format) {
		contentType = "text/xml"
		res, err = tools.XmlUncamel(data)
	} else {
		contentType = "application/json; charset=utf-8"
		//res, err = json.Marshal(data)
		res, err = safejson.MarshalSafeCollections(data)
	}
	if err != nil {
		h.handlerError(w, errors.New("返回数据解析失败."), ctx, false)
		return
	}

	// 响应
	w.Header().Set("Content-Type", contentType)
	func() {
		// 解决透传其它openapi错误码的问题。当返回非http类型错误码时，会产生panic，在此进行一次转换屏蔽
		defer func() {
			if r := recover(); r != nil {
				logger.Error("%s Invalid http code %d", ctx.requestId, ctx.status)
				w.WriteHeader(400)
			}
		}()
		w.WriteHeader(ctx.status)
	}()
	w.Write(res)
}
