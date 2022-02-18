package queue

import (
	"container/heap"
	"github.com/Zoxu0928/task-common/logger"

	"sync"
	"time"
)

// 延迟队列

type DealyQueue struct {
	name   string
	queue  PriorityQueue // 优先队列
	lock   sync.Mutex    // 队列操作锁
	wait   *time.Timer   // 当队列中无数据，或首个元素延迟时间未到时，需要Wait
	active bool          // false代表Queue已关闭
}

func NewDealyQueue(name string) *DealyQueue {
	return &DealyQueue{
		name:   name,
		queue:  make(PriorityQueue, 0, 32),
		wait:   time.NewTimer(time.Second * 30),
		active: true,
	}
}

// 增加元素
func (this *DealyQueue) Add(dItem *DealyItem) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if dItem == nil {
		return false
	}
	if this.active == false {
		return false
	}
	heap.Push(&this.queue, &dItem.item)
	peek := this.queue.Peek()
	if peek != nil && peek == &dItem.item {
		this.wait.Reset(dItem.GetDelay())
	}
	return true
}

// 删除元素
func (this *DealyQueue) Remove(dItem *DealyItem) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if dItem == nil {
		return false
	}
	if this.active == false {
		return false
	}
	if dItem.item.index < 0 {
		return false
	}
	heap.Remove(&this.queue, dItem.item.index)
	return true
}

// 更新元素
func (this *DealyQueue) Update(dItem *DealyItem, value interface{}, dealy time.Time) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if dItem == nil {
		return false
	}
	if this.active == false {
		return false
	}
	if dItem.item.index < 0 {
		return false
	}
	this.queue.update(&dItem.item, value, dealy.UnixNano())
	peek := this.queue.Peek()
	if peek != nil && peek == &dItem.item {
		this.wait.Reset(dItem.GetDelay())
	}
	return true
}

// timer线程不安全，reset时需要同步操作

// 获取第一个到期元素，如果取不到，会阻塞在此
// 返回false代表延迟队列已关闭
func (this *DealyQueue) Take() (dealyItem DealyItem, ok bool) {

	for {

		func() {

			// 加锁
			this.lock.Lock()
			defer this.lock.Unlock()

			// 延期时间
			var dealy time.Duration

			// 获取队列中第一个元素，并未真正取出。
			// 如果没有取到，说明队列是空的，等待30秒
			// 取到了，计算延期时间，时间小于0代表已到期。大于0代表未到期，未到期就按延期时间进行等待
			if this.active == false {
				dealy = 0
			} else if first := this.queue.Peek(); first == nil {
				dealy = time.Second * 30
			} else {
				dealy = time.Duration(first.priority - time.Now().UnixNano())
				if dealy <= time.Microsecond {
					dealyItem, ok = DealyItem{item: *heap.Pop(&this.queue).(*Item)}, true
				}
			}

			// 重置延期时间
			this.wait.Reset(dealy)
		}()

		// 已关闭
		if this.active == false {
			return
		}

		// 取到了则返回，取不到则等待
		if ok {
			return
		} else {
			<-this.wait.C
		}
	}
}

// 队列大小
func (this *DealyQueue) Size() int {
	return len(this.queue)
}

// 关闭队列
func (this *DealyQueue) Close() {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.active {
		logger.Info("[DealyQueue close] %s", this.name)
		this.queue.Clear()
		this.queue = nil
		this.active = false
		this.wait.Reset(0) // 唤醒其它线程
	}
}

// 延迟队列元素

type DealyItem struct {
	item Item
}

func NewDealyItem(value interface{}, dealy time.Time) DealyItem {
	return DealyItem{
		item: Item{
			value:    value,
			priority: dealy.UnixNano(),
		},
	}
}

// 获得延期时间，小于等于0时表示已到期
func (this DealyItem) GetDelay() time.Duration {
	return time.Duration(this.item.priority - time.Now().UnixNano())
}

func (this DealyItem) GetValue() interface{} {
	return this.item.value
}
