package main

import "sync"

// StringQueue is a thread-safe queue for string type
type StringQueue struct {
	storage []string
	sync.Mutex
}

// NewStringQueue returns a new StringQueue pointer
func NewStringQueue() *StringQueue {
	return &StringQueue{
		storage: []string{},
	}
}

// Push appends the given string to the tail of the StringQueue
func (s *StringQueue) Push(ss string) {
	s.Lock()
	defer s.Unlock()
	s.storage = append(s.storage, ss)
}

// Pop returns the element at the head of the StringQueue and removes it from the StringQueue
func (s *StringQueue) Pop() string {
	s.Lock()
	defer s.Unlock()
	var a string
	a, s.storage = s.storage[0], s.storage[1:]

	return a
}

// Len returns the current length of the StringQueue
func (s *StringQueue) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.storage)
}
