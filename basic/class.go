package basic

//**************************************
// 类
//**************************************

import (
	"reflect"
	"strings"
	"sync"
)

// 声明类
type Class struct {
	mu      sync.RWMutex       //锁
	cls     interface{}        //目标对象
	name    string             //类名称
	methods map[string]*Method //方法集合
}

// 实例化
func NewClass(cls interface{}) *Class {
	return createClass(cls)
}

//**************************************
// 公开方法
//**************************************
// 获得类名称
func (this *Class) GetName() string {
	return this.name
}

// 根据方法名称获得方法，如：pkg.funcname
func (this *Class) GetMethod(methodName string) *Method {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.methods[methodName]
}

// 获得此类中所有方法
func (this *Class) GetAllMethod() map[string]*Method {
	return this.methods
}

//**************************************
// 私有方法
//**************************************
func createClass(cls interface{}) *Class {

	// 实例化
	obj := &Class{
		methods: make(map[string]*Method),
		cls:     cls,
		name:    GetClassName(cls),
	}

	// 遍历所有方法并实例化并缓存起来
	ele := reflect.TypeOf(cls)
	if ele.NumMethod() == 0 {
		return obj
	}

	for i := 0; i < ele.NumMethod(); i++ {
		obj.addMethod(NewMethod(obj, ele.Method(i)))
	}
	return obj
}

func (this *Class) addMethod(method *Method) {
	if method == nil {
		return
	}
	this.mu.Lock()
	defer this.mu.Unlock()
	this.methods[method.GetName()] = method
}

//**************************************
// 对外静态方法
//**************************************
func GetClassName(v interface{}) string {
	return strings.Replace(reflect.TypeOf(v).String(), "*", "", 1)
}
