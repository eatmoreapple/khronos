package khronos

import (
	"container/heap"
	"context"
	"sync"
)

func PqFromContext(ctx context.Context) *PriorityQueueWithRouting {
	return ctx.Value(QueueContextKey).(*PriorityQueueWithRouting)
}

func PqWithContext(ctx context.Context, pq *PriorityQueueWithRouting) context.Context {
	return context.WithValue(ctx, QueueContextKey, pq)
}

// Item represents an item in the queue.
type Item struct {
	value    string // The value of the item.
	priority int64  // The priority of the item.
	index    int    // The index of the item in the heap.
}

// PriorityQueue implements a priority queue.
type PriorityQueue []*Item

// Len returns the length of the priority queue.
func (pq PriorityQueue) Len() int { return len(pq) }

// Less compares two items by their priorities.
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

// Swap swaps two items in the priority queue.
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the priority queue.
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the item with the highest priority from the priority queue.
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// PriorityQueueWithRouting implements a thread-safe priority queue with routing support.
type PriorityQueueWithRouting struct {
	queueMap  map[string]*PriorityQueue // Map of queues based on routes.
	queueLock sync.Mutex                // Lock for concurrent access to the queues.
	notEmpty  map[string]*sync.Cond     // Condition variables for each route to block when the queue is empty.
}

// NewPriorityQueueWithRouting creates a new instance of PriorityQueueWithRouting.
func NewPriorityQueueWithRouting() *PriorityQueueWithRouting {
	return &PriorityQueueWithRouting{
		queueMap: make(map[string]*PriorityQueue),
		notEmpty: make(map[string]*sync.Cond),
	}
}

// Enqueue adds an item to the queue based on the specified route and priority.
func (pq *PriorityQueueWithRouting) Enqueue(route string, item *Item) {
	pq.queueLock.Lock()
	defer pq.queueLock.Unlock()

	queue, ok := pq.queueMap[route]
	if !ok {
		queue = &PriorityQueue{}
		heap.Init(queue)
		pq.queueMap[route] = queue
	}

	heap.Push(queue, item)

	cond, condExists := pq.notEmpty[route]
	if !condExists {
		cond = sync.NewCond(&pq.queueLock)
		pq.notEmpty[route] = cond
	}

	cond.Broadcast()
}

// Dequeue removes and returns the item with the highest priority from the queue based on the specified route.
// If the queue is empty, it blocks until an item is available.
func (pq *PriorityQueueWithRouting) Dequeue(route string) *Item {
	pq.queueLock.Lock()

	for {
		queue, ok := pq.queueMap[route]
		if ok && queue.Len() > 0 {
			item := heap.Pop(queue).(*Item)
			pq.queueLock.Unlock()
			return item
		}

		cond, condExists := pq.notEmpty[route]
		if !condExists {
			cond = sync.NewCond(&sync.Mutex{})
			pq.notEmpty[route] = cond
		}

		pq.queueLock.Unlock() // 释放主锁，允许其他队列操作
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
		pq.queueLock.Lock() // 重新获取主锁
	}
}

func (pq *PriorityQueueWithRouting) Length(route string) int {
	pq.queueLock.Lock()
	defer pq.queueLock.Unlock()

	queue, ok := pq.queueMap[route]
	if !ok {
		return 0
	}
	return queue.Len()
}
