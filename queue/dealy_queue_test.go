package queue

import (
	"container/heap"
	"fmt"
	"github.com/Zoxu0928/task-common/basic"

	"math/rand"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

// 一次性消费
func TestDealyQueueConsume1(t *testing.T) {

	// 写入一百万数据，全部1秒后到期
	total := 1000000
	fmt.Println(basic.Time(time.Now()), "准备数据：", total, "个")
	dq := NewDealyQueue("测试")
	for i := 1; i <= total; i++ {
		item := NewDealyItem(strconv.Itoa(i), time.Now().Add(time.Second*1))
		dq.Add(&item)
	}

	// 并发take测试结果:
	// 消费线程并发越大耗时会越多，争抢严重，只需要单线程take即可。如果需要控制并发，那么在take到数据之后另起线程控制即可。
	start := time.Now()
	paral := 1
	fmt.Println(basic.Time(time.Now()), "消费线程：", paral, "个")
	count := int32(0)
	for i := 1; i <= paral; i++ {
		go func(num int) {
			for {
				_, ok := dq.Take()
				if ok {
					atomic.AddInt32(&count, 1)
					//fmt.Println(fmt.Sprintf("%s 协程 %d 处理延迟元素 %v", basic.Time(time.Now()), num, item.GetValue()))
				} else {
					fmt.Println(fmt.Sprintf("%s 队列关闭 -> 协程 %d 退出", basic.Time(time.Now()), num))
					return
				}
			}
		}(i)
	}

	for {
		c := atomic.LoadInt32(&count)
		fmt.Println(basic.Time(time.Now()), "已处理", c)
		if c < int32(total) {
			time.Sleep(time.Millisecond * 100)
		} else {
			break
		}
	}
	fmt.Println(basic.Time(time.Now()), "处理完毕，耗时->", time.Since(start))

	fmt.Println(basic.Time(time.Now()), "关闭队列")
	dq.Close()

	time.Sleep(time.Second)
}

// 持续消费
func TestDealyQueueConsume2(t *testing.T) {

	// 写入一百万数据，全部1秒后到期
	total := 1000000
	fmt.Println(basic.Time(time.Now()), "准备数据：", total, "个")
	dq := NewDealyQueue("测试")
	for i := 1; i <= total; i++ {
		item := NewDealyItem(strconv.Itoa(i), time.Now().Add(time.Second*1))
		dq.Add(&item)
	}

	// 并发take测试结果:
	// 消费线程并发越大耗时会越多，争抢严重，只需要单线程take即可。如果需要控制并发，那么在take到数据之后另起线程控制即可。
	paral := 1
	fmt.Println(basic.Time(time.Now()), "消费线程：", paral, "个")
	count := int32(0)
	for i := 1; i <= paral; i++ {
		go func(num int) {
			for {
				_, ok := dq.Take()
				if ok {
					atomic.AddInt32(&count, 1)
					//fmt.Println(fmt.Sprintf("%s 协程 %d 处理延迟元素 %v", basic.Time(time.Now()), num, item.GetValue()))
				} else {
					fmt.Println(fmt.Sprintf("%s 队列关闭 -> 协程 %d 退出", basic.Time(time.Now()), num))
					return
				}
			}
		}(i)
	}

	// 另起线程写入数据
	for i := 1; i <= 10; i++ {
		go func() {
			for {
				time.Sleep(time.Second * 5)
				add := 100000
				fmt.Println(basic.Time(time.Now()), "追加元素", add, "个")
				for n := 1; n <= add; n++ {
					item := NewDealyItem(strconv.Itoa(i), time.Now().Add(time.Second*1))
					dq.Add(&item)
				}
			}
		}()
	}

	for {
		c := atomic.LoadInt32(&count)
		fmt.Println(basic.Time(time.Now()), "已处理", c)
		time.Sleep(time.Millisecond * 200)
	}
}

// 测试删除
func TestDealyQueueRemove(t *testing.T) {

	var del1 *DealyItem
	var del2 *DealyItem

	total := 10
	fmt.Println(basic.Time(time.Now()), "准备数据：", total, "个")
	dq := NewDealyQueue("测试")
	for i := 1; i <= total; i++ {
		item := NewDealyItem(strconv.Itoa(i), time.Now().Add(time.Second*time.Duration(rand.Intn(10))))
		dq.Add(&item)
		if i == 3 {
			del1 = &item
		}
		if i == 7 {
			del2 = &item
		}
	}
	fmt.Println("当前队列:")
	Print(&dq.queue)

	// 删除二个
	fmt.Println("删除", dq.Remove(del1))
	Print(&dq.queue)
	fmt.Println("删除", dq.Remove(del2))
	Print(&dq.queue)

	for dq.queue.Len() > 0 {
		item := heap.Pop(&dq.queue).(*Item)
		fmt.Printf("取出 %.2d: %s \n", (item.priority-time.Now().UnixNano())/1000000000, item.value)
		Print(&dq.queue)
	}
}

// 测试更新
func TestDealyQueueUpdate(t *testing.T) {

	var update1 *DealyItem
	var update2 *DealyItem

	total := 10
	fmt.Println(basic.Time(time.Now()), "准备数据：", total, "个")
	dq := NewDealyQueue("测试")
	for i := 1; i <= total; i++ {
		item := NewDealyItem(strconv.Itoa(i), time.Now().Add(time.Second*time.Duration(rand.Intn(10))))
		dq.Add(&item)
		if i == 3 {
			update1 = &item
		}
		if i == 7 {
			update2 = &item
		}
	}
	fmt.Println("当前队列:")
	Print(&dq.queue)

	fmt.Println("更新", update1)
	dq.Update(update1, "333333", time.Now().Add(100*time.Second))
	Print(&dq.queue)

	fmt.Println("更新", update2)
	dq.Update(update2, "777777", time.Now().Add(-2*time.Second))
	Print(&dq.queue)

	fmt.Println("更新", &DealyItem{})
	dq.Update(update2, "555555", time.Now().Add(2*time.Second))
	Print(&dq.queue)

	for dq.queue.Len() > 0 {
		item := heap.Pop(&dq.queue).(*Item)
		fmt.Printf("取出 %.2d: %s \n", (item.priority-time.Now().UnixNano())/1000000000, item.value)
		Print(&dq.queue)
	}
}

func Print(pq *PriorityQueue) {
	old := *pq
	for i := 0; i < old.Len(); i++ {
		item := old[i]
		fmt.Printf("%d - %.2d - %s \n", item.index, (item.priority-time.Now().UnixNano())/1000000000, item.value)
	}
	fmt.Println()
}
