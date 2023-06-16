package khronos

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func ExamplePriorityQueue() {
	pq := NewPriorityQueueWithRouting()

	// Enqueue items with different routes and priorities.
	pq.Enqueue("route", &Item{value: "item2", priority: 2})
	pq.Enqueue("route", &Item{value: "item3", priority: 3})
	pq.Enqueue("route", &Item{value: "item1", priority: 1})

	// Dequeue items based on routes.
	item1 := pq.Dequeue("route")
	fmt.Println(item1.value) // Output: item1
}

func BenchmarkPriorityQueue(b *testing.B) {
	pq := NewPriorityQueueWithRouting()

	for i := 0; i < b.N; i++ {
		pq.Enqueue("route1", &Item{value: "item1", priority: 1})
		pq.Enqueue("route2", &Item{value: "item2", priority: 2})
		pq.Enqueue("route1", &Item{value: "item3", priority: 3})
		pq.Enqueue("route2", &Item{value: "item4", priority: 4})
		pq.Dequeue("route1")
		pq.Dequeue("route2")
	}
	// BenchmarkPriorityQueue-8   	 2592499	       494.9 ns/op
}

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueueWithRouting()

	// Enqueue items with different routes and priorities.
	pq.Enqueue("route", &Item{value: "item2", priority: 2})
	pq.Enqueue("route", &Item{value: "item3", priority: 3})
	pq.Enqueue("route", &Item{value: "item1", priority: 1})

	// Dequeue items based on routes.
	item1 := pq.Dequeue("route")
	if item1.value != "item1" {
		t.Errorf("Expected item1, got %s", item1.value)
	}
	item2 := pq.Dequeue("route")
	if item2.value != "item2" {
		t.Errorf("Expected item2, got %s", item2.value)
	}
	item3 := pq.Dequeue("route")
	if item3.value != "item3" {
		t.Errorf("Expected item3, got %s", item3.value)
	}
}

func TestPriorityQueue_Pop(t *testing.T) {
	pq := NewPriorityQueueWithRouting()

	go func() {
		time.Sleep(1 * time.Second)
		pq.Enqueue("route", &Item{value: "item2", priority: 2})
	}()

	item1 := pq.Dequeue("route")
	if item1.value != "item2" {
		t.Errorf("Expected item2, got %s", item1.value)
	}
}

func benchmarkEnqueueDequeue(b *testing.B, numWorkers int) {
	pq := NewPriorityQueueWithRouting()

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()

			item := &Item{
				value:    "example",
				priority: 0,
			}

			for j := 0; j < b.N; j++ {
				pq.Enqueue("route", item)
				pq.Dequeue("route")
			}
		}()
	}

	wg.Wait()
}

func BenchmarkEnqueueDequeue4(b *testing.B) {
	benchmarkEnqueueDequeue(b, 4)
	// BenchmarkEnqueueDequeue4-8   	 1232782	      1016 ns/op
}

func BenchmarkEnqueueDequeue10(b *testing.B) {
	benchmarkEnqueueDequeue(b, 10)
	// BenchmarkEnqueueDequeue10-8   	  465942	      2583 ns/op
}

func BenchmarkEnqueueDequeue100(b *testing.B) {
	benchmarkEnqueueDequeue(b, 100)
	// BenchmarkEnqueueDequeue100-8   	   44734	     27689 ns/op
}

func BenchmarkEnqueueDequeue1000(b *testing.B) {
	benchmarkEnqueueDequeue(b, 1000)
	// BenchmarkEnqueueDequeue1000-8   	    4462	    270369 ns/op
}
