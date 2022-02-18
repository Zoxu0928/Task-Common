package global

import (
	"github.com/Zoxu0928/task-common/logger"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"
)

const (
	STATE_INIT = iota
	STATE_RUNNING
	STATE_SHUTTING_DOWN
	STATE_TERMINATE
)

// 全局资源管理

// 创建任何资源，只要是需要Close的资源，全部要加入到资源管理中
// 例如：
//  kafka连接
//  mysql连接
//  redis连接
//  zookeeper连接
//  网络socket资源
//  web服务
//  其它自定义的资源，比如有守护线程之类的资源

var DefaultHammerTime time.Duration
var DisableGracefullyStopped bool

// Resource 所有需要关闭的资源，都要实现的接口
type Resource interface {
	Close()
}

var DefaultResourceManager = &resourceManager{
	wg:              sync.WaitGroup{},
	mu:              &sync.RWMutex{},
	resourcesBefore: make([]Resource, 0),
	resources:       make([]Resource, 0),
	resourcesAfter:  make([]Resource, 0),
	state:           STATE_INIT,
}

type resourceManager struct {
	wg              sync.WaitGroup
	mu              *sync.RWMutex
	resourcesBefore []Resource
	resources       []Resource
	resourcesAfter  []Resource
	state           uint8
}

// Add 添加一个需要释放的资源（先释放）
func (srv *resourceManager) AddBeforeFree(res Resource) {
	logger.Info("add resource - [%s]", reflect.TypeOf(res).Elem().Name())
	srv.resourcesBefore = append(srv.resourcesBefore, res)
}

// Add 添加一个需要释放的资源（后释放；web 服务专属）
func (srv *resourceManager) AddAfterFree(res Resource) {
	logger.Info("add resource - [%s]", reflect.TypeOf(res).Elem().Name())
	srv.setState(STATE_RUNNING)
	srv.resourcesAfter = append(srv.resourcesAfter, res)
}

// Add 添加一个需要释放的资源（中释放）
func (srv *resourceManager) Add(res Resource) {
	logger.Info("add resource - [%s]", reflect.TypeOf(res).Elem().Name())
	srv.resources = append(srv.resources, res)
}

func (srv *resourceManager) setState(st uint8) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	srv.state = st
}

func (srv *resourceManager) GetState() uint8 {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	return srv.state
}

// 销毁所有资源
func (srv *resourceManager) destroy() {
	if srv.GetState() != STATE_RUNNING {
		srv.wg.Wait()
		return
	}
	srv.wg.Add(1)
	srv.setState(STATE_SHUTTING_DOWN)
	srv.resourcesBefore = srv.eachCloseResources(srv.resourcesBefore)

	if DefaultHammerTime >= 0 {
		logger.Info("destroy wait %s", DefaultHammerTime)
		time.Sleep(DefaultHammerTime)
	}

	srv.resources = srv.eachCloseResources(srv.resources)
	srv.setState(STATE_TERMINATE)
	srv.wg.Done()
	logger.Info("destroy success!")
}

// 关闭资源
func (srv *resourceManager) eachCloseResources(list []Resource) []Resource {
	for _, v := range list {
		if v != nil {
			// TODO: 这里是否需要做 recover ？
			v.Close()
			logger.Info("offline resource - [%s]", reflect.TypeOf(v).Elem().Name())
		}
	}
	return make([]Resource, 0)
}

// ListenSignal 监听Kill -15 信号
func (srv *resourceManager) ListenSignal() {
	sigs := make(chan os.Signal, 1)
	defer func() {
		signal.Stop(sigs)
		close(sigs)
	}()

Loop:
	for {
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigs
		logger.Info("Accept signal %s.", sig)
		if DisableGracefullyStopped == true {
			break Loop
		}
		switch sig {
		case syscall.SIGINT:
			// kill -2（通过此信号销毁 web 服务以外的资源）
			logger.Info("Received syscall SIGINT.")
			srv.destroy()

		case syscall.SIGTERM:
			// kill -15（系统、K8s 默认使用此信号终止 Pod；走这里才关闭 web 服务）
			logger.Info("Received syscall SIGTERM.")
			srv.destroy()
			srv.resourcesAfter = srv.eachCloseResources(srv.resourcesAfter)
			logger.Info("The application is shutting down...")
			break Loop
		default:
			// 监听的信号都被处理，理论上不走这
			logger.Info("Received %v: nothing i care about...\n", sig)
		}
	}
}
