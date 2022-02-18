package queue

import (
	"container/heap"
)

// 优先队列，利用 Golang Heap 实现

// An Item is something we manage in a priority queue.
type Item struct {
	value    interface{} // The value of the item; arbitrary.
	priority int64       // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value interface{}, priority int64) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// 获得数组中第一个元素，因为数组中第一个元素一定是最优先的。得到该元素但并不从数组中删除。pop操作才是真正的拿出并删除。
func (pq *PriorityQueue) Peek() *Item {
	old := *pq
	n := len(old)
	if n <= 0 {
		return nil
	}
	return old[0]
}

func (pq *PriorityQueue) Clear() {
	pq = nil
}
