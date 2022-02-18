package controller

//**************************************
// 路由接口
// 通过AddController将一个实体类、或一个Func增加到路由接口中
// 通过GetController可以获取某个接口method，再通过返回的method.Invoke执行目标函数
//**************************************

import (
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/bean"
	"github.com/Zoxu0928/task-common/logger"
	"reflect"
)

// 增加controller
func (this *controller) AddController(v interface{}) {
	if v == nil {
		logger.Error("增加Controller失败，参数不能为空")
		return
	}
	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}
	if vType.Kind() == reflect.Struct {
		this.addClass(v)
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Warn("[[[[[[[[[[ Add bean from controller failed. controller=%s ]]]]]]]]]]", basic.GetClassName(v))
				}
			}()
			bean.AddBean(v) // Controller也属于bean，放入bean缓存
		}()
	} else if vType.Kind() == reflect.Func {
		this.addMethod(v)
	}
}

// 根据方法名称获得方法
func (this *controller) GetController(methodName, version string) *basic.Method {
	this.mmu.RLock()
	defer this.mmu.RUnlock()
	return this.methods[this.getControllerKey(version, methodName)]
}

// 向全局路由中增加controller
func AddController(v interface{}) {
	default_controller.AddController(v)
}
