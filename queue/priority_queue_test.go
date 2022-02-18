package queue

import (
	"container/heap"
	"fmt"
	"testing"
)

// This example creates a PriorityQueue with some items, adds and manipulates an item,
// and then removes the items in priority order.
func TestPriorityQueue(t *testing.T) {
	// Some items and their priorities.
	items := map[string]int64{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		fmt.Println(i, value, priority)
		i++
	}

	heap.Init(&pq)

	// Insert a new item and then modify its priority.
	item := &Item{
		value:    "orange",
		priority: 1,
	}

	heap.Push(&pq, item)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d: %d: %s \n", item.priority, item.index, item.value)
	}
}
