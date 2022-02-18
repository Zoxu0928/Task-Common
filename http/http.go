package http

import (
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/e"
	"net/http"
)

// 接口
type HttpClient interface {

	// Get请求，返回string
	Get(requestId, url string, header map[string]string) (string, e.ApiError)

	// Get请求，输入request struct，返回值注入response struct中
	GetJson(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError

	// Post请求，输入request struct，返回值注入response struct中
	PostJson(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError

	// Post请求，输入request map，返回值注入response struct中
	PostForm(requestId, url string, form map[string]interface{}, response interface{}, header map[string]string) e.ApiError

	// Put请求
	Put(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError

	// Patch请求
	Patch(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError

	// Delete请求
	Delete(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError

	// 转发
	RoundTripPost(requestId, url string, header map[string][]string, reqData []byte) (resBody []byte, resStatus int, err e.ApiError)
	RoundTrip(requestId, url string, r *http.Request) (resBody []byte, resStatus int, err e.ApiError)

	// 原生
	GetTransport() *http.Transport
}

// 配置
type HttpClientConf struct {
	TimeOut             basic.Duration `yaml:"timeout" toml:"timeout"`
	MaxIdleConns        int            `yaml:"max_idle_conns" toml:"max_idle_conns"`
	MaxIdleConnsPerHost int            `yaml:"max_idle_conns_per_host" toml:"max_idle_conns_per_host"`
}
