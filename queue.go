package main

import "sync"

// StringQueue is a thread-safe work queue for string type. It tracks
// pending items, which are items in the queue plus items popped but not
// yet marked Done, so that Pop can tell a temporarily empty queue from a
// finished one.
type StringQueue struct {
	storage []string
	pending int
	closed  bool
	cond    *sync.Cond
	sync.Mutex
}

// NewStringQueue returns a new StringQueue pointer
func NewStringQueue() *StringQueue {
	s := &StringQueue{
		storage: []string{},
	}
	s.cond = sync.NewCond(&s.Mutex)
	return s
}

// Push appends the given string to the tail of the StringQueue
func (s *StringQueue) Push(ss string) {
	s.Lock()
	defer s.Unlock()
	if s.closed {
		return
	}
	s.storage = append(s.storage, ss)
	s.pending++
	s.cond.Signal()
}

// Pop returns the element at the head of the StringQueue and removes it from
// the StringQueue. It blocks while the StringQueue is empty and other popped
// items are not yet Done. It returns false when no work remains or the
// StringQueue is closed.
func (s *StringQueue) Pop() (string, bool) {
	s.Lock()
	defer s.Unlock()
	for len(s.storage) == 0 && s.pending > 0 && !s.closed {
		s.cond.Wait()
	}
	if s.closed || len(s.storage) == 0 {
		return "", false
	}
	var a string
	a, s.storage = s.storage[0], s.storage[1:]
	return a, true
}

// Done marks one popped element as fully processed
func (s *StringQueue) Done() {
	s.Lock()
	defer s.Unlock()
	s.pending--
	if s.pending == 0 {
		s.cond.Broadcast()
	}
}

// Close makes all current and future calls to Pop return false
func (s *StringQueue) Close() {
	s.Lock()
	defer s.Unlock()
	s.closed = true
	s.cond.Broadcast()
}
