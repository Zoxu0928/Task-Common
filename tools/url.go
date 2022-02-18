package tools

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// 生成请求地址，不带参数部分
func GetUrl(r *http.Request) string {
	schema := r.URL.Scheme
	host := r.Host
	path := r.URL.Path
	if schema == "" {
		schema = "http"
	}
	return fmt.Sprintf("%s://%s%s", schema, host, path)
}

// 根据地址，获取其中的参数
func GetUrlParam(urlStr, key string) string {
	ul, _ := url.Parse(urlStr)
	if ul != nil {
		for k, v := range ul.Query() {
			if strings.EqualFold(k, key) && len(v) > 0 {
				return v[0]
			}
		}
	}
	return ""
}
