package interceptor

// 用途：注册拦截器并在请求进入时依次执行匹配的拦截器
// 调用次序：
//     执行目标方法之前，按拦截器加入的顺序，依次执行BeforeFunc
//     执行目标方法之后，按拦截器加入的顺序，反向依次执行AfterFunc
//     在进入拦截器之后，拦截器返回之前，在此时间内出现错误，将按反方向依次执行(运行过的拦截器)OnError，如果在调用第2个拦截器时出错，那么会按2->1的顺序执行此方法，并不会执行所有的拦截器
//
// 拦截器中途退出的时机
//     1. 当执行某个拦截器出错时，流程将结束并返回错误
//     2. 当执行某个拦截器（必须是非daemon的拦截器）之后，ReqContext中的Response字段被赋值了，流程将结束并正常返回

import (
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/bean"
	"github.com/Zoxu0928/task-common/cache"
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"
	"reflect"
	"strings"
)

// 增加一个拦截器
func (this *interceptor) Add(pointCut *PointCut) {
	if pointCut.CutFunc == nil {
		return
	}
	logger.Info("add Interceptor -> Name=%s Target=%s Matchs=%s Exclude=%s Daemon=%t", pointCut.Name, strings.Replace(reflect.TypeOf(pointCut.CutFunc).Elem().String(), "*", "", 1), pointCut.Matchs, pointCut.Excludes, pointCut.Daemon)
	this.chain = append(this.chain, pointCut)
	func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Warn("[[[[[[[[[[ Add bean from interceptor failed. interceptor=%s ]]]]]]]]]]", pointCut.Name)
			}
		}()
		bean.AddBean(pointCut.CutFunc) // 拦截器也属于bean，放入bean缓存
	}()
}

// 判断拦截器是否匹配目标函数
// 匹配参考：
// method.GetName())    jcs.DescribeInstances
func (pCut *PointCut) isMatch(method *basic.Method) bool {

	// 是否排除
	for _, excludeStr := range pCut.Excludes {

		// * 为匹配所有拦截器
		if excludeStr == "*" {
			return false

			// 当目标全路径，与拦截器规则一致时
		} else if excludeStr == method.GetFullName() {
			return false

			// 当目标路径(不带名称)，与拦截器规则一致时
		} else if excludeStr == method.GetPath() {
			return false

			// 当目标名称(不带路径)，与拦截器规则一致时
		} else if excludeStr == method.GetName() {
			return false

			// 判断拦截器规则路径是否模糊匹配目标路径
			// a/b/c/* 可以匹配 a/b/c/d/e/ecs.CreateInstance
		} else if tools.MatchPath(method.GetFullName(), excludeStr) {
			return false
		}
	}

	match := false

	// 是否匹配
	for _, matchStr := range pCut.Matchs {

		// * 为匹配所有拦截器
		if matchStr == "*" {
			match = true

			// 当目标全路径，与拦截器规则一致时
		} else if matchStr == method.GetFullName() {
			match = true

			// 当目标路径(不带名称)，与拦截器规则一致时
		} else if matchStr == method.GetPath() {
			match = true

			// 当目标名称(不带路径)，与拦截器规则一致时
		} else if matchStr == method.GetName() {
			match = true

			// 判断拦截器规则路径是否模糊匹配目标路径
			// a/b/c/* 可以匹配 a/b/c/d/e/ecs.CreateInstance
		} else if tools.MatchPath(method.GetFullName(), matchStr) {
			match = true
		}

		if match {
			break
		}
	}

	return match
}

// 执行拦截器
func (this *interceptor) DoIntercepptor(ctx Context, method *basic.Method, doFinal func() error) error {

	// 保存目标方法上应该执行的拦截器列表
	var cuts []*PointCut
	cache_key := "interceptor." + method.GetFullName()

	// 偿试从缓存中获取目标方法需要执行的拦截器
	cache_value, _ := cache.Local.Get(cache_key)
	if cache_value != nil {
		cuts = cache_value.([]*PointCut)
	} else {

		// 计算匹配的拦截器
		cuts = make([]*PointCut, 0)
		for _, pCut := range this.chain {
			if pCut.isMatch(method) {
				cuts = append(cuts, pCut)
			}
		}
		// 加入缓存
		cache.Local.Set(cache_key, cuts, -1)

		var getNames = func() []string {
			s := make([]string, 0)
			for _, v := range cuts {
				s = append(s, v.Name)
			}
			return s
		}
		logger.Info("Interceptor match result: %s %s", method.GetFullName(), getNames())
	}

	// 错误
	var globalError error = nil

	// 当前执行到第几个拦截器
	index := 0

	// Before
	if len(cuts) > 0 {
		for ; index < len(cuts); index++ {
			pCut := cuts[index]
			if pCut.Daemon {
				go func() {
					defer e.OnError(ctx.GetRequestId())
					pCut.CutFunc.BeforeFunc(ctx)
				}()
			} else {
				globalError = pCut.CutFunc.BeforeFunc(ctx)
				if globalError != nil {
					break
				}
				// 在拦截器中就已经得到了响应内容，那么流程到此就可以结束了，原路返回
				if ctx.GetResponse() != nil {
					break
				}
			}
		}
	}

	// 所有拦截器没有报错，并且此时Response为空
	// 执行目标
	if globalError == nil && ctx.GetResponse() == nil {
		globalError = doFinal()
	}

	if index >= len(cuts) {
		index = len(cuts) - 1
	}

	// After，反向调用
	if globalError == nil && len(cuts) > 0 {
		for ; index >= 0; index-- {
			pCut := cuts[index]
			if pCut.Daemon {
				go func() {
					defer e.OnError(ctx.GetRequestId())
					pCut.CutFunc.AfterFunc(ctx)
				}()
			} else {
				globalError = pCut.CutFunc.AfterFunc(ctx)
				if globalError != nil {
					break
				}
			}
		}
	}

	// 如果过程有错，从当前位置一层层返向通知被执行过的拦截器
	if globalError != nil {
		for ; index >= 0; index-- {
			pCut := cuts[index]
			if pCut.Daemon {
				go func() {
					defer e.OnError(ctx.GetRequestId())
					pCut.CutFunc.OnError(ctx, globalError)
				}()
			} else {
				pCut.CutFunc.OnError(ctx, globalError)
			}
		}
	}

	return globalError
}
