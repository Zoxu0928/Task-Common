package cache

// 本地缓存

import (
	"github.com/Zoxu0928/task-common/e"
	"github.com/Zoxu0928/task-common/global"
	"github.com/Zoxu0928/task-common/logger"
	"github.com/Zoxu0928/task-common/queue"
	"github.com/Zoxu0928/task-common/tools"
	"regexp"
	"runtime/debug"
	"sync"
	"time"
)

// 默认创建一个本地缓存
var Local = CrateLocalCache()

// 声明类
type localCache struct {
	mu         sync.RWMutex
	cache      map[string]*lcc
	dealyQueue *queue.DealyQueue
}

// 缓存结构体
type lcc struct {
	key       string
	value     interface{}
	dealyItem *queue.DealyItem
}

// 实例化
func CrateLocalCache() *localCache {
	lc := &localCache{
		cache:      make(map[string]*lcc),
		dealyQueue: queue.NewDealyQueue("default local cache"),
	}
	global.DefaultResourceManager.Add(lc.dealyQueue)
	go lc.monitorKey() // 启动一个协程处理到期key的清理工作
	return lc
}

// 监控到期key并清理
func (this *localCache) monitorKey() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Got a runtime error. %s\n%s", r, string(debug.Stack()))
			go this.monitorKey()
		}
	}()
	for {
		dealyItem, ok := this.dealyQueue.Take() // ok=false表示队列已关闭
		if ok {
			go func() {
				defer e.OnError("clear expired key")
				key := dealyItem.GetValue()
				this.remove_expired(key.(string))
			}()
		} else {
			return
		}
	}
}

// 实现接口

func (this *localCache) Set(k string, v interface{}, period time.Duration) error {
	_, err := this.setNx(k, v, period, false)
	if err != nil {
		return err
	}
	return nil
}

func (this *localCache) SetNx(k string, v interface{}, period time.Duration) (int, error) {
	return this.setNx(k, v, period, true)
}

func (this *localCache) Get(k string) (interface{}, error) {
	logger.Debug("LocalCache Get -> %s", k)
	this.mu.RLock()
	defer this.mu.RUnlock()
	data := this.cache[k]
	if data == nil {
		return nil, nil
	}
	return data.value, nil
}

func (this *localCache) Remove(k string) error {
	logger.Debug("LocalCache Remove -> %s", k)
	this.mu.Lock()
	defer this.mu.Unlock()
	data := this.cache[k]
	delete(this.cache, k)
	if data != nil {
		this.dealyQueue.Remove(data.dealyItem)
	}
	return nil
}

func (this *localCache) remove_expired(k string) error {
	logger.Debug("LocalCache Expired Remove -> %s", k)
	this.mu.Lock()
	defer this.mu.Unlock()
	delete(this.cache, k)
	return nil
}

func (this *localCache) Time(k string) (int, error) {
	logger.Debug("LocalCache Time -> %s", k)
	this.mu.RLock()
	defer this.mu.RUnlock()
	data := this.cache[k]
	if data == nil || data.dealyItem == nil {
		return -1, nil
	}
	leftTime := int(data.dealyItem.GetDelay() / 1000000000)
	if leftTime < 0 {
		return 0, nil
	}
	return leftTime, nil
}

func (this *localCache) Size() (int, error) {
	logger.Debug("dealy queue size %d", this.dealyQueue.Size())
	return len(this.cache), nil
}

func (this *localCache) setNx(k string, v interface{}, period time.Duration, nx bool) (int, error) {

	// 上锁
	this.mu.Lock()
	defer this.mu.Unlock()

	// 从缓存中取出数据
	data := this.cache[k]

	// 已存在，返回
	if nx && data != nil {
		return -1, nil
	}

	logger.Debug("LocalCache Put -> %s", k)

	// 如果缓存中没有，则创建一个新的
	// 如果有，更新value
	if data == nil {
		data = &lcc{key: k, value: v}
		this.cache[k] = data
	} else {
		data.value = v
	}

	// 如果过期时间大于0
	if period > 0 {
		if data.dealyItem == nil {
			dealyItem := queue.NewDealyItem(k, time.Now().Add(period)) // 第一次加入缓存
			data.dealyItem = &dealyItem
			this.dealyQueue.Add(&dealyItem)
		} else {
			this.dealyQueue.Update(data.dealyItem, k, time.Now().Add(period)) // 已经存在，只需要更新
		}
	} else {
		if data.dealyItem != nil {
			this.dealyQueue.Remove(data.dealyItem)
			data.dealyItem = nil
		}
	}

	return 1, nil
}

func (this *localCache) Close() {
	if this.dealyQueue != nil {
		this.dealyQueue.Close()
	}
}

func (this *localCache) Mon() string {
	return ""
}

// 筛选Key
func (this *localCache) Keys(key string) ([]string, error) {

	arr := []string{}

	if key == "" {
		return arr, nil
	}

	// 正则
	reg, err := regexp.Compile(tools.ToRegStr(key, true))
	if err != nil {
		return arr, err
	}

	// 匹配
	for k, _ := range this.cache {
		if key == "*" || reg.Match([]byte(k)) {
			arr = append(arr, k)
		}
	}

	return arr, nil
}
