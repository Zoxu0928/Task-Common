package interceptor

import (
	"github.com/Zoxu0928/task-common/basic"
	"net/http"
)

//---------------------------------------------
// 拦截器中传递的上下文，具体需要业务方自己实现里面的内容
type Context interface {
	SetRequest(model interface{})             //接口入参
	GetRequest() interface{}                  //接口入参
	SetResponse(model interface{})            //接口返回
	GetResponse() interface{}                 //接口返回
	GetRequestId() string                     //请求ID
	GetHttpRequest() *http.Request            //获取http请求
	GetHeader(key string) string              //获取http请求的header
	GetAttribute(key string) interface{}      //获取缓存属性
	SetAttribute(key string, val interface{}) //设置缓存属性
	GetModel() string
	GetAction() string
}

// 定义拦截器结构 begin
// 拦截器属性
type PointCut struct {
	Name     string   //拦截器名称
	Daemon   bool     //拦截器是否异步执行
	Matchs   []string //拦截器需要匹配的路径
	Excludes []string //拦截器需要排除的路径
	CutFunc  CutFunc  //拦截器执行的方法
}

// 拦截器需要实现的接口
type CutFunc interface {
	BeforeFunc(ctx Context) error   //调用目标之前，按拦截器加入的顺序，依次执行此方法
	AfterFunc(ctx Context) error    //调用目标之后，按拦截器加入的顺序，反向依次执行此方法
	OnError(ctx Context, err error) //在进入拦截器之后，拦截器返回之前，在此时间内出现错误，将按反方向依次执行(运行过的拦截器)此方法，如果在调用第2个拦截器时出错，那么会按2->1的顺序执行此方法，并不会执行所有的拦截器
}

// 实例化拦截器属性
func NewPointCut(name string, matchs []string, excludes []string, daemon bool) *PointCut {
	return &PointCut{
		Name:     name,
		Daemon:   daemon,
		Matchs:   matchs,
		Excludes: excludes,
	}
}

// 定义拦截器结构 end
//---------------------------------------------

// 抽象接口
type Interceptor interface {
	Add(pointCut *PointCut)
	DoIntercepptor(ctx Context, method *basic.Method, doFinal func() error) error
}

// 拦截器链表
type interceptor struct {
	chain []*PointCut
}

// 创建一个新的拦截器链表
func NewInterceptpr() *interceptor {
	return &interceptor{
		chain: make([]*PointCut, 0),
	}
}

// 默认初始化一个全局拦截器链表
var default_interceptor = NewInterceptpr()

// 获得全局拦截器链表
func GetDefaultInterceptor() *interceptor {
	return default_interceptor
}

// 获得全局拦截器链表
func GetInterceptor() *interceptor {
	return default_interceptor
}
