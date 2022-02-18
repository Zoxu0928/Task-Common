package api

import (
	"net/http"
	"strconv"
)

// 所有请求struct的父struct
type Request struct {
	// 每个请求对应唯一的ID
	RequestId string
	// 用戶
	Account string
	// 子帐户
	User string
	// 地域，具体校验从配置中取
	RegionId string
	// 业务模块
	Model string
	// 请求的具体方法
	Action string
	// 接口返回格式，支持xml和json，默认采用json格式渲染响应体
	Format string
	// 获取http上下文
	GetHttpContext func() (http.ResponseWriter, *http.Request) `json:"-"`
	// 请求来源
	BusiCode string `json:",omitempty"`
	// 运营后台请求帐号
	ErpAccount string `json:",omitempty"`
	// 角色，一般如果调用方是代入角色替用户操作的，网关会将角色在Header中传过来
	Role string `json:",omitempty"`
	// 用户ID，在用户验证拦截器中设置
	tenant_id string
}

// 请求是否来自前端控制台
//func (this *Request) IsConsoleRequest() bool {
//	_, r := this.GetHttpContext()
//	if r == nil {
//		return false
//	}
//	return r.Header.Get(constants.HEADER_CONSOLE) == "true" || this.BusiCode == constants.BusiCode.Console
//}

// 所有响应struct的父struct
type Response struct {
	Message string `json:"message"`
}

// 过滤条件
type Filter struct {
	Name     string   `json:"name"`
	Operator string   `json:"operator"`
	Values   []string `json:"values"`
}

func (f Filter) GetPointerValue() []*string {
	res := make([]*string, len(f.Values))
	for i, _ := range f.Values {
		res[i] = &f.Values[i]
	}
	return res
}

type TagFilter struct {
	Key      string
	Operator string
	Values   []string
}

type FilterGroup struct {
	Filters []*Filter
}

func (this *Request) SetTenantId(tid string) {
	this.tenant_id = tid
}

func (this *Request) GetTenantId() string {
	return this.tenant_id
}

// 增加ResponseHeader
func (this *Request) AddResponseHeader(k, v string) {
	if w, _ := this.GetHttpContext(); w != nil {
		w.Header().Set(k, v)
	}
}

type Filters []*Filter

func (f Filters) GetFilterValuesByKey(key string) interface{} {
	for _, filter := range f {
		if filter.Name == key {
			return filter.Values
		}
	}
	return nil
}

func (f Filters) GetFilterPointerValuesByKey(key string, tt string) interface{} {
	for _, filter := range f {
		if filter.Name == key {
			switch tt {
			case "string":
				res := make([]*string, len(filter.Values))
				for i, _ := range filter.Values {
					res[i] = &filter.Values[i]
				}
				return res
			case "int32":
				res := make([]*int32, len(filter.Values))
				for i, _ := range filter.Values {
					v, _ := strconv.ParseInt(filter.Values[i], 10, 32)
					w := int32(v)
					res[i] = &w
				}
				return res
			}
			return nil
		}
	}
	return nil
}
