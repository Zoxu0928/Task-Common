package controller

import (
	"fmt"
	"github.com/Zoxu0928/task-common/basic"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/tools"

	"sync"
)

// 抽象接口
type Controller interface {
	AddController(v interface{})
	GetController(methodName, version string) *basic.Method
	UseModel() bool
}

// 声明类
type controller struct {
	cmu      sync.RWMutex             //类锁
	mmu      sync.RWMutex             //方法锁
	useModel bool                     //方法key格式开关
	classes  map[string]*basic.Class  //类集合
	methods  map[string]*basic.Method //方法集合。useModel=true时，key=版本.model.funcName。useModel=false时，key=版本.funcName。
}

// 实例化
func NewController(useModel bool) *controller {
	return &controller{
		classes:  make(map[string]*basic.Class),
		methods:  make(map[string]*basic.Method),
		useModel: useModel,
	}
}

//**************************************
// 私有方法
//**************************************

// 缓存一个类
func (this *controller) addClass(v interface{}) {

	// 获得类名称
	className := basic.GetClassName(v)

	// 从缓存中获得类
	class := this.getClass(className)

	// 如果缓存中不存在，则创建一个
	if class == nil {

		// 创建成功，将类与方法集合分别缓存起来
		class = basic.NewClass(v)
		if class != nil {
			this.putClass(class)
			this.putAllMethod(class)
		} else {
			panic(fmt.Sprintf("add controller error. class is nil. className=%s", className))
		}
	}
}

// 缓存一个方法，此方法不属于类，只是一个单纯的Func
func (this *controller) addMethod(v interface{}) {
	methodName := basic.GetMethodName(v)
	method := basic.NewMethod(nil, v)
	if method != nil {
		this.putMethod(method)
	} else {
		panic(fmt.Sprintf("add controller error. method is nil. methodName=%s", methodName))
	}
}

// 根据类名称，获得一个类信息
func (this *controller) getClass(className string) *basic.Class {
	this.cmu.RLock()
	defer this.cmu.RUnlock()
	return this.classes[className]
}

// 设置一个类信息
func (this *controller) putClass(class *basic.Class) {
	this.cmu.Lock()
	defer this.cmu.Unlock()
	this.classes[class.GetName()] = class
}

// 设置一个类中的所有方法
func (this *controller) putAllMethod(class *basic.Class) {
	this.mmu.Lock()
	defer this.mmu.Unlock()
	for name, v := range class.GetAllMethod() {
		version := getVersion(v.GetPath())
		logger.Info("add controller -> Name=[%s] Version=[%s] Path=[%s] useModel=[%t]", name, version, v.GetPath(), this.useModel)
		key := this.getControllerKey(version, name)
		if _, ok := this.methods[key]; ok {
			panic("Found Duplicate controller name " + key)
		}
		this.methods[key] = v
	}
}

// 设置一个方法
func (this *controller) putMethod(method *basic.Method) {
	this.mmu.Lock()
	defer this.mmu.Unlock()
	version := getVersion(method.GetPath())
	logger.Info("add controller -> Name=[%s] Version=[%s] Path=[%s] useModel=[%t]", method.GetName(), version, method.GetPath(), this.useModel)
	key := this.getControllerKey(version, method.GetName())
	if v := this.methods[key]; v == nil {
		if _, ok := this.methods[key]; ok {
			panic("Found Duplicate controller name " + key)
		}
		this.methods[key] = method
	}
}

// 从路径上获得version值
func getVersion(path string) string {
	if tools.EndWith(path, "/") {
		return tools.Replace(tools.Replace(path, "/$", ""), ".*/", "")
	} else {
		return tools.Replace(path, ".*/", "")
	}
}

// 生成controller的key
func (this *controller) getControllerKey(version, name string) string {
	key := version + "."
	if this.useModel {
		key = key + name
	} else {
		key = key + tools.Replace(name, `^[^.]*[.]`, "")
	}
	return key
}

func (this *controller) UseModel() bool {
	return this.useModel
}

// 默认初始化一个全局使用的路由
var default_controller = NewController(false)

// 获得全局默认的路由
func GetDefaultController() *controller {
	return default_controller
}
