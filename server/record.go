package main

import (
	"time"
)

type Record[T any] struct {
	put       chan<- T
	get       chan<- chan Entry[T]
	getRecent chan<- chan Entry[T]
}

type Entry[T any] struct {
	t time.Time
	v T
}

// Create a record with the specified capacity.
// If the capacity is exceeded, old entires will be discarded and new ones kept.
func newRecord[T any](capacity int) Record[T] {
	put := make(chan T)
	get := make(chan chan Entry[T])
	getRecent := make(chan chan Entry[T])

	go func() {
		entries := make([]Entry[T], 0, capacity)

		for {
			select {
			case v, ok := <-put:
				if !ok {
					return
				}
				entries = append(entries, Entry[T]{time.Now(), v})
				if len(entries) > capacity {
					entries = entries[1:]
				}
			case c, ok := <-get:
				if !ok {
					return
				}
				for _, e := range entries {
					c <- e
				}
				close(c)
			case c, ok := <-getRecent:
				if !ok {
					return
				}
				if len(entries) > 0 {
					c <- entries[len(entries)-1]
				}
				close(c)
			}
		}
	}()

	return Record[T]{put, get, getRecent}
}

func (l Record[T]) Close() {
	close(l.put)
	close(l.get)
	close(l.getRecent)
}
