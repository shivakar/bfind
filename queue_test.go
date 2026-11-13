package main

import (
	"strconv"
	"sync"
	"testing"
)

// TestStringQueueConcurrentTraversal walks a synthetic tree from a single root
// with many workers, which is the case where workers used to race on an
// almost empty queue.
func TestStringQueueConcurrentTraversal(t *testing.T) {
	const depth, fanout, workers = 6, 3, 16

	for run := 0; run < 50; run++ {
		q := NewStringQueue()
		q.Push("")

		var mu sync.Mutex
		seen := 0
		var wg sync.WaitGroup
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					node, ok := q.Pop()
					if !ok {
						return
					}
					mu.Lock()
					seen++
					mu.Unlock()
					if len(node) < depth {
						for c := 0; c < fanout; c++ {
							q.Push(node + strconv.Itoa(c))
						}
					}
					q.Done()
				}
			}()
		}
		wg.Wait()

		want := 0
		for d, n := 0, 1; d <= depth; d, n = d+1, n*fanout {
			want += n
		}
		if seen != want {
			t.Fatalf("run %d: visited %d nodes, want %d", run, seen, want)
		}
	}
}

func TestStringQueueClose(t *testing.T) {
	q := NewStringQueue()
	q.Push("a")
	if _, ok := q.Pop(); !ok {
		t.Fatal("Pop returned false on a non-empty queue")
	}

	popped := make(chan bool)
	go func() {
		_, ok := q.Pop()
		popped <- ok
	}()
	q.Close()
	if ok := <-popped; ok {
		t.Fatal("Pop returned true after Close")
	}
}
