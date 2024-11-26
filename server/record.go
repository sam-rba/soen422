package main

import (
	"time"
)

type Record[T any] struct {
	put       chan<- T
	get       chan<- chan T
	getRecent chan<- chan T
}

type entry[T any] struct {
	t time.Time
	v T
}

// Create a record with the specified capacity.
// If the capacity is exceeded, old entires will be discarded and new ones kept.
func newRecord[T any](capacity int) Record[T] {
	put := make(chan T)
	get := make(chan chan T)
	getRecent := make(chan chan T)

	go func() {
		entries := make([]entry[T], 0, capacity)

		for {
			select {
			case v, ok := <-put:
				if !ok {
					return
				}
				entries = append(entries, entry[T]{time.Now(), v})
				if len(entries) > capacity {
					entries = entries[1:]
				}
			case c, ok := <-get:
				if !ok {
					return
				}
				for _, e := range entries {
					c <- e.v
				}
				close(c)
			case c, ok := <-getRecent:
				if !ok {
					return
				}
				if len(entries) > 0 {
					c <- entries[len(entries)-1].v
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
