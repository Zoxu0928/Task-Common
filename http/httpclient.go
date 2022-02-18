package http

import (
	"bytes"
	"errors"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"
	"github.com/Zoxu0928/task-common/tools/json/ffjson"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unsafe"
)

const (
	GET       = "GET"
	POST      = "POST"
	JSON      = "application/json;charset=UTF-8"
	FORM      = "application/x-www-form-urlencoded"
	STATUS_OK = 200
)

// 声明类
type httpClient struct {
	timeOut   time.Duration
	transport *http.Transport
}

// 实例化
func CreateHttpClient(conf *HttpClientConf) *httpClient {
	time_out := 20 * time.Second
	max_idle_conns := 1000
	max_idle_conns_per_host := 30
	if conf != nil && conf.TimeOut.Duration > 0 {
		time_out = conf.TimeOut.Duration
	}
	if conf != nil && conf.MaxIdleConns > 0 {
		max_idle_conns = conf.MaxIdleConns
	}
	if conf != nil && conf.MaxIdleConnsPerHost > 0 {
		max_idle_conns_per_host = conf.MaxIdleConnsPerHost
	}
	return &httpClient{
		timeOut: time_out,
		transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          max_idle_conns,
			MaxIdleConnsPerHost:   max_idle_conns_per_host,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

// 实现接口
func (this *httpClient) GetTransport() *http.Transport {
	return this.transport
}

// get请求
func (this *httpClient) Get(requestId, url string, header map[string]string) (string, e.ApiError) {

	logger.Info("%s GET[NORMAL] URL=%s Headers=%s", requestId, url, tools.ToString(header))

	// 创建Request
	request, err := http.NewRequest(GET, url, nil)
	if err != nil {
		return "", httpError(requestId, "创建RemoteRequest失败", err)
	}

	// 设置Header
	for k, v := range header {
		request.Header.Set(k, v)
	}

	// 发出请求
	client := &http.Client{Timeout: this.timeOut, Transport: this.transport}
	remoteResponse, err := client.Do(request)
	if err != nil {
		return "", httpError("Http远程访问出错", requestId, err)
	}

	// 返回
	return httpResponse(requestId, remoteResponse, err)
}

// PostForm请求
func (this *httpClient) PostForm(requestId, _url string, form map[string]interface{}, response interface{}, header map[string]string) e.ApiError {

	logger.Info("%s POST[FORM] URL=%s Headers=%s Param=%s", requestId, _url, tools.ToString(header), form)

	values := url.Values{}
	for k, v := range form {
		if str_val, ok := v.(string); ok {
			values.Add(k, str_val)
		} else {
			values.Add(k, tools.ToString(v))
		}
	}

	// 创建Request
	request, err := http.NewRequest(POST, _url, strings.NewReader(values.Encode()))
	if err != nil {
		return httpError(requestId, "创建RemoteRequest失败", err)
	}

	// 设置Header
	request.Header.Set("Content-Type", FORM)
	for k, v := range header {
		request.Header.Set(k, v)
	}

	// 发出请求
	client := &http.Client{Timeout: this.timeOut, Transport: this.transport}
	remoteResponse, err := client.Do(request)
	if err != nil {
		return httpError("Http远程访问出错", requestId, err)
	}

	// 响应
	data, apiErr := httpResponse(requestId, remoteResponse, err)
	if apiErr != nil {
		ffjson.Unmarshal([]byte(data), response)
		return apiErr
	}

	if err := ffjson.Unmarshal([]byte(data), response); err != nil {
		return httpError("解析响应数据失败.", requestId, err)
	}

	return nil
}

// Get请求
func (this *httpClient) GetJson(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {
	return this.do(http.MethodGet, requestId, url, request, response, header)
}

// Post请求
func (this *httpClient) PostJson(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {
	return this.do(http.MethodPost, requestId, url, request, response, header)
}

// Put请求
func (this *httpClient) Put(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {
	return this.do(http.MethodPut, requestId, url, request, response, header)
}

// Patch请求
func (this *httpClient) Patch(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {
	return this.do(http.MethodPatch, requestId, url, request, response, header)
}

// Delete请求
func (this *httpClient) Delete(requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {
	return this.do(http.MethodDelete, requestId, url, request, response, header)
}

// 执行http请求
func (this *httpClient) do(method string, requestId, url string, request interface{}, response interface{}, header map[string]string) e.ApiError {

	// 解析json
	jsonByte, err := tools.Marshal(request)
	if err != nil {
		return httpError(requestId, "参数解析失败", err)
	}

	jsonStr := string(jsonByte)

	logger.Info("%s %s URL=%s Headers=%s Param=%s", requestId, method, url, tools.ToString(header), jsonStr)

	// 创建Request
	var reader io.Reader = nil
	if request != nil {
		reader = bytes.NewReader(jsonByte)
	}
	httpRequest := new(http.Request)
	_httpRequest, err := http.NewRequest(method, url, reader)
	if err != nil {
		return httpError(requestId, "创建RemoteRequest失败", err)
	}
	*httpRequest = *_httpRequest

	// 设置Header
	httpRequest.Header.Set("Content-Type", JSON)
	for k, v := range header {
		httpRequest.Header.Set(k, v)
	}

	// 发出请求
	client := &http.Client{Timeout: this.timeOut, Transport: this.transport}
	var remoteResponse *http.Response
	var respErr error
	// 如果失败重试两次
	for i := 1; i <= 3; i++ {
		remoteResponse, respErr = client.Do(httpRequest)
		if respErr != nil {
			logger.Error("Http远程访问出错, error: %s", respErr.Error())
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if respErr != nil {
		return httpError("Http远程访问出错", requestId, respErr)
	}

	// 响应
	data, apiErr := httpResponse(requestId, remoteResponse, err)
	if apiErr != nil {
		if response != nil {
			ffjson.Unmarshal([]byte(data), response)
		}
		return apiErr
	}

	if response != nil {
		if err := ffjson.Unmarshal([]byte(data), response); err != nil {
			baseDataStr := "{ \"data\":" + data + "}"
			if err := ffjson.Unmarshal([]byte(baseDataStr), response); err != nil {
				return httpError("解析响应数据失败.", requestId, err)
			}
		}

	}

	// Nova返回的token临时保存，Nova下线时删掉此处代码
	if header != nil {
		if token := remoteResponse.Header.Get("X-Subject-Token"); token != "" {
			header["X-Subject-Token"] = token
		}
	}

	return nil
}

// 处理响应
func httpResponse(requestId string, response *http.Response, err error) (string, e.ApiError) {

	if response != nil {
		defer response.Body.Close()
	}

	// 请求失败
	if err != nil {
		return "", httpError("Http远程访问出错", requestId, err)
	}

	// 解析响应
	body, err := ioutil.ReadAll(response.Body)

	// 解析失败
	if err != nil {
		return "", httpError("Http响应数据解析失败", requestId, err)
	}

	// 响应结果
	content := (*string)(unsafe.Pointer(&body))

	logger.Info("%s HttpResponse(%s) %s", requestId, tools.ToString(response.StatusCode), *content)

	// http状态不是200
	if response.StatusCode != STATUS_OK && response.StatusCode != 201 && response.StatusCode != 204 {

		// 404类错误
		if response.StatusCode == 404 {
			return *content, e.NewApiError(e.NOT_FOUND, "not found.", nil)
		}

		// 尝试将底层错误转换为rpcError
		var rpcErr *rpcError
		if *content != "" {
			rpcErr = &rpcError{}
			ffjson.Unmarshal(body, rpcErr)
		}

		// 如果成功转换，原样返回
		if rpcErr != nil && rpcErr.Code != "" {
			return *content, e.NewApiError(e.INTERNAL, rpcErr.Code+" - "+rpcErr.Message, errors.New(*content))
		} else {
			return *content, e.NewApiError(e.UNKNOWN, "Http响应错误:"+tools.ToString(response.StatusCode), errors.New(*content))
		}
	}

	return *content, nil
}

// 统一处理错误
func httpError(msg, requestId string, cause error) e.ApiError {
	if cause != nil {
		logger.Error("%s %s err=%s", requestId, msg, cause)
	} else {
		logger.Error("%s %s", requestId, msg)
	}
	if ce, ok := cause.(e.ApiError); ok {
		return ce
	} else {
		return e.NewApiError(e.UNKNOWN, msg, cause)
	}
}

type rpcError struct {
	Code, Message string
}

func (this *httpClient) Close() {
}

// 转发POST请求
func (this *httpClient) RoundTripPost(requestId, url string, header map[string][]string, reqData []byte) (resBody []byte, resStatus int, err e.ApiError) {
	return this.roundTripNewBody(POST, requestId, url, header, reqData)
}

// 转发请求，使用新request body
func (this *httpClient) roundTripNewBody(method, requestId, url string, header map[string][]string, reqData []byte) (resBody []byte, resStatus int, err e.ApiError) {

	logger.Info("%s %s URL=%s Headers=%s Param=%s", requestId, method, url, tools.ToString(header), string(reqData))

	// 创建Request
	var reader io.Reader = nil
	if reqData != nil {
		reader = bytes.NewReader(reqData)
	}
	httpRequest, reqErr := http.NewRequest(method, url, reader)
	if reqErr != nil {
		err = httpError(requestId, "创建RemoteRequest失败", reqErr)
		return
	}

	// 设置Header
	httpRequest.Header.Set("Content-Type", JSON)
	for k, v := range header {
		for _, d := range v {
			httpRequest.Header.Add(k, d)
		}
	}

	return this.RoundTrip(requestId, url, httpRequest)
}

// 转发请求，使用源request
func (this *httpClient) RoundTrip(requestId, rawurl string, r *http.Request) (resBody []byte, resStatus int, err e.ApiError) {

	logger.Info("RoundTrip %s URL=%s", requestId, rawurl)

	if r.URL.Scheme == "" {
		rUrl, pErr := url.Parse(rawurl)
		if pErr != nil {
			logger.Error("%s ParseUrl Error. %s", requestId, pErr)
			return
		}
		r.URL = rUrl
	}

	// 请求
	response, tripErr := this.transport.RoundTrip(r)

	// 释放资源
	if response != nil && response.Body != nil {
		defer response.Body.Close()
	}
	if tripErr != nil {
		err = httpError(requestId, "调用RoundTrip失败", tripErr)
		return
	}

	// 获取响应数据
	body, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		err = httpError(requestId, "读取ResponseBody失败", readErr)
		return
	}

	logger.Info("%s HttpResponse(%s) %s", requestId, tools.ToString(response.StatusCode), string(body))

	resBody = body
	resStatus = response.StatusCode
	return
}
