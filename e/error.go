package e

import (
	"errors"
	"github.com/Zoxu0928/task-common/tools"
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"github.com/Zoxu0928/task-common/tools/json/jsonparser"

	"strings"
)

// Api错误接口
type ApiError interface {
	GetCode() int
	GetType() string
	GetMessage() string
	SetDetails(details []map[string]string)
	GetDetails() []map[string]string
	AddDetail(k, v string)
	GetCause() error
	Error() string
	IsNotFound() bool
	IsHttpError() bool
	SetOErrCode(eCode string)
	GetOErrCode() string
}

// Api错误类
type busiError struct {
	Code     int
	Type     string
	Message  string
	Cause    error
	Details  []map[string]string
	oerrCode string
}

// 实现接口
func (e *busiError) GetCode() int {
	return e.Code
}
func (e *busiError) GetType() string {
	return e.Type
}
func (e *busiError) GetMessage() string {
	return e.Message
}
func (e *busiError) SetDetails(details []map[string]string) {
	e.Details = details
}
func (e *busiError) AddDetail(k, v string) {
	if e.Details == nil {
		e.Details = make([]map[string]string, 1, 1)
		e.Details[0] = map[string]string{}
	}
	e.Details[0][k] = v
}
func (e *busiError) GetDetails() []map[string]string {
	return e.Details
}
func (e *busiError) GetCause() error {
	if e.Cause == nil {
		e.Cause = errors.New("")
	}
	return e.Cause
}
func (e *busiError) Error() string {
	return tools.ToJson(e)
}

// 设置交易的错误Code
func (e *busiError) SetOErrCode(eCode string) {
	e.oerrCode = eCode
}
func (e *busiError) GetOErrCode() string {
	return e.oerrCode
}

// 是否是NotFound类型的错误
func (e *busiError) IsNotFound() bool {
	if e.GetCode() == NOT_FOUND.Code {
		return true
	}
	// 硬盘网络
	if strings.Index(e.Error(), "notfound") > -1 {
		return true
	}
	if strings.Index(e.Error(), "doesn't belong") > -1 {
		return true
	}
	// 主机
	if strings.Index(e.Error(), "not_found") > -1 {
		return true
	}
	return false
}

// 是否是网络错误
func (e *busiError) IsHttpError() bool {
	if e.GetCode() == 500 && strings.Contains(e.GetMessage(), "Http远程") {
		return true
	}
	return false
}

// 是否数据库唯一冲突错误
func IsDuplicateEntry(err error) bool {
	return strings.Index(err.Error(), "Error 1062") > -1
}

// 实例化
func NewApiError(errorCode *ErrorCode, message string, cause error) ApiError {
	return &busiError{
		Code:    errorCode.Code,
		Type:    errorCode.Type,
		Message: message,
		Cause:   convertNormalError(cause),
	}
}

// 实例化
func NewNormalApiError(code int, Type, message string, cause error) ApiError {
	return &busiError{
		Code:    code,
		Type:    Type,
		Message: message,
		Cause:   convertNormalError(cause),
	}
}

// 实例化
func UnmarshalApiError(text string) ApiError {

	if text == "" {
		return nil
	}

	// 解析json字符串，计算错误结构的嵌套层次及Cause字段的类型应该是什么
	var prepare func(data *[]byte, be *busiError)
	prepare = func(data *[]byte, be *busiError) {

		CauseVal, Type, _, _ := jsonparser.Get(*data, "cause")
		switch Type {
		case jsonparser.Object:
			if _, e1 := jsonparser.GetInt(CauseVal, "code"); e1 == nil {
				cause := &busiError{}
				be.Cause = cause
				prepare(&CauseVal, cause)
			} else if _, e2 := jsonparser.GetString(CauseVal, "err"); e2 == nil {
				be.Cause = &ie{}
			} else {
				be.Cause = errors.New("")
			}
		case jsonparser.String:
		case jsonparser.Null:
		}
	}

	// 保存结果的结构体
	busiErr := &busiError{}

	// 字符串转成字节
	data_bytes := []byte(text)

	// 计算结构体的嵌套层次
	prepare(&data_bytes, busiErr)

	// json解析
	if err := ffjson.Unmarshal([]byte(text), busiErr); err != nil {
		return UnknownError(err)
	} else {
		return busiErr
	}
}

// 实例化:内部错误
func InternalError(cause error) ApiError {
	return &busiError{
		Code:    INTERNAL.Code,
		Type:    INTERNAL.Type,
		Message: "Internal server error",
		Cause:   convertNormalError(cause),
	}
}

// 实例化:未知错误
func UnknownError(cause error) ApiError {
	return &busiError{
		Code:    UNKNOWN.Code,
		Type:    UNKNOWN.Type,
		Message: "Unknown server error",
		Cause:   convertNormalError(cause),
	}
}

// NOT_FOUND 类型错误
func NotFoundError(msg string, cause error) ApiError {
	return &busiError{
		Code:    NOT_FOUND.Code,
		Type:    NOT_FOUND.Type,
		Message: msg,
		Cause:   cause,
	}
}

// 将普通error转换为我的error->ie
func convertNormalError(err error) error {
	if err == nil {
		return err
	}
	switch err.(type) {
	case ApiError:
		return err
	default:
		return &ie{err.Error()}
	}
}

// 我的错误，首字母大写，保证tojson的时候可以得到信息
type ie struct {
	Err string
}

func (e *ie) Error() string {
	return e.Err
}
