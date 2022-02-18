package basic

import (
	"errors"
	"fmt"
	"sync"
)

// 链表结构
type linkedList struct {
	lock  sync.Mutex // 协程安全的链表
	size  int        // 链表大小
	first *node      // 链表中第一个元素
	last  *node      // 链表中最后一个元素
}

// 链表中的元素结构
type node struct {
	data interface{} // 当前数据
	prev *node       // 上一个元素
	next *node       // 下一个元素
}

// 创建一个空链表
func NewLinkedList() *linkedList {
	return &linkedList{}
}

// 追加一个元素
func (this *linkedList) Add(v interface{}) {
	this.AddLast(v)
}

// 在最后追加一个元素
func (this *linkedList) AddLast(v interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()

	// 更换最后一个元素
	currentLastNode := this.last
	newLastNode := &node{data: v, prev: currentLastNode, next: nil}
	this.last = newLastNode

	// 如果当前没有最后一个元素，那么头和尾都指向一个即可
	if currentLastNode == nil {
		this.first = newLastNode
	} else {
		currentLastNode.next = newLastNode
	}
	this.size = this.size + 1
}

// 在开始位置追加一个元素
func (this *linkedList) AddFirst(v interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()

	// 更换第一个元素
	currentFirstNode := this.first
	newFirstNode := &node{data: v, prev: nil, next: currentFirstNode}
	this.first = newFirstNode

	// 如果当前没有第一个元素，那么头和尾指向一个即可
	if currentFirstNode == nil {
		this.last = newFirstNode
	} else {
		currentFirstNode.prev = newFirstNode
	}
	this.size = this.size + 1
}

// 在指定位置前追加一个元素，下标从0开始
func (this *linkedList) AddBefore(index int, v interface{}) error {
	if index < 0 || index > this.size {
		return errors.New("linkedList.AddBefore: index out of range")
	}
	if index == this.size {
		this.AddLast(v)
	} else {
		this.lock.Lock()
		defer this.lock.Unlock()
		targetNode := this.getNode(index)
		if targetNode == nil {
			return errors.New(fmt.Sprintf("item of '%d' not found", index))
		} else {
			currentPrevNode := targetNode.prev
			newPrevNode := &node{data: v, prev: currentPrevNode, next: targetNode}
			targetNode.prev = newPrevNode
			if currentPrevNode == nil {
				this.first = newPrevNode
			} else {
				currentPrevNode.next = newPrevNode
			}
			this.size = this.size + 1
		}
	}
	return nil
}

// 删除一个元素
func (this *linkedList) Remove(index int) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	if index < 0 || index >= this.size {
		return nil
	}
	targetNode := this.getNode(index)
	return this.deleteNode(targetNode)
}

// 删除指定数据的元素
func (this *linkedList) RemoveData(v interface{}) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	var targetNode *node
	if v == nil {
		for n := this.first; n != nil; n = n.next {
			if n.data == nil {
				targetNode = n
				break
			}
		}
	} else {
		for n := this.first; n != nil; n = n.next {
			if n.data == v {
				targetNode = n
				break
			}
		}
	}
	if targetNode != nil {
		this.deleteNode(targetNode)
		return true
	}
	return false
}

// 从前向后开始删除，包含指定下标
func (this *linkedList) RemoveUntil(index int) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if index < 0 || index >= this.size {
		return errors.New("linkedList.RemoveUntil: index out of range")
	}
	current_index := 0
	for n := this.first; n != nil; {
		next := n.next
		n.data = nil
		n.prev = nil
		n.next = nil
		this.first = next
		this.size = this.size - 1

		if next == nil {
			this.last = nil
		} else {
			next.prev = nil
		}

		if current_index == index {
			n = nil
			break
		} else {
			current_index = current_index + 1
			n = next
		}
	}
	return nil
}

// 删除头
func (this *linkedList) RemoveFirst() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.first == nil {
		return nil
	}
	currentFirstNode := this.first
	dat := currentFirstNode.data

	newFirstNode := currentFirstNode.next
	this.first = newFirstNode
	if newFirstNode == nil {
		this.last = nil
	} else {
		newFirstNode.prev = nil
	}

	currentFirstNode.data = nil
	currentFirstNode.prev = nil
	currentFirstNode.next = nil

	this.size = this.size - 1

	return dat
}

// 删除尾
func (this *linkedList) RemoveLast() interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.last == nil {
		return nil
	}
	currentLastNode := this.last
	dat := currentLastNode.data

	newLastNode := currentLastNode.prev
	this.last = newLastNode
	if newLastNode == nil {
		this.first = nil
	} else {
		newLastNode.next = nil
	}

	currentLastNode.data = nil
	currentLastNode.prev = nil
	currentLastNode.next = nil

	this.size = this.size - 1

	return dat
}

// 链表大小
func (this *linkedList) Size() int {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.size
}

// 清空链表
func (this *linkedList) Clear() {
	this.lock.Lock()
	defer this.lock.Unlock()
	for n := this.first; n != nil; {
		next := n.next
		n.data = nil
		n.prev = nil
		n.next = nil
		n = next
	}
	this.first = nil
	this.last = nil
	this.size = 0
}

// 获取指定下标数据
func (this *linkedList) Get(index int) interface{} {
	this.lock.Lock()
	defer this.lock.Unlock()
	if index < 0 || index >= this.size {
		return nil
	}
	return this.getNode(index).data
}

// 遍历链表，返回true继续，返回false停止
func (this *linkedList) Loop(callback func(index int, v interface{}) bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	index := 0
	for n := this.first; n != nil; n = n.next {
		if n == nil {
			break
		}
		if callback(index, n.data) {
			index++
		} else {
			break
		}
	}
}

// 查找指定位置的元素
func (this *linkedList) getNode(no int) *node {

	var n *node

	// 优化查找顺序，是从前开始找，还是从后开始找
	if no < (this.size / 2) {
		n = this.first
		for from := 0; from < no; from++ {
			n = n.next
		}
	} else {
		n = this.last
		for from := this.size - 1; from > no; from-- {
			n = n.prev
		}
	}

	return n
}

// 删除指定元素
func (this *linkedList) deleteNode(n *node) interface{} {

	if n == nil {
		return nil
	}

	dat := n.data

	t_prevNode := n.prev
	t_NextNode := n.next

	if t_prevNode == nil {
		this.first = t_NextNode
	} else {
		t_prevNode.next = t_NextNode
	}

	if t_NextNode == nil {
		this.last = t_prevNode
	} else {
		t_NextNode.prev = t_prevNode
	}

	n.data = nil
	n.prev = nil
	n.next = nil

	this.size = this.size - 1

	return dat
}
