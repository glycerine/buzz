package buzz

import (
	"fmt"
	"math/rand"
	"sync"
)

// SyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to SyncTower.
//
// Sends always block until all receives have happened.
// Subscribers are given unbuffered channels.
//
type SyncTower struct {
	subscribers map[string]chan interface{}
	names       []string
	mu          sync.Mutex
}

func NewSyncTower() *SyncTower {
	return &SyncTower{
		subscribers: make(map[string]chan interface{}),
	}
}

func (b *SyncTower) Subscribe(name string) chan interface{} {
	b.mu.Lock()
	ch := make(chan interface{}, 1)
	b.subscribers[name] = ch
	b.names = append(b.names, name)
	b.mu.Unlock()
	return ch
}

func (b *SyncTower) Unsub(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subscribers, name)

	// names is separate, also needs updating
	k := -1
	for i := range b.names {
		if b.names[i] == name {
			k = i
			break
		}
	}
	if k == -1 {
		return fmt.Errorf("name not subscribed: '%s'", name)
	}
	// delete names[k] from the middle of names
	n := len(b.names)
	switch {
	case k == 0:
		if n == 1 {
			b.names = nil
		} else {
			b.names = b.names[1:]
		}

	case k < n-1: // k >= 1 and n >= 2
		b.names = append(b.names[:k], b.names[(k+1):]...)

	case k == n-1:
		b.names = b.names[:(n - 1)]
	}
	return nil
}

// Broadcast sends a copy of val to all subscribers.
// Since the receivers are all unbuffered
// channels, Broadcast will  block
// waiting on all receives.
//
// Any subscriber who subscribes after the Broadcast will not
// receive the Broadcast value, as it is not
// stored internally.
//
func (b *SyncTower) Broadcast(val interface{}) {
	b.mu.Lock()
	for _, ch := range b.subscribers {
		select {
		case ch <- val:
		default:
		}
	}
	b.mu.Unlock()
}

// Signal works like sync.Cond's Signal.
// It sends val to exactly one listener.
//
// The listener is chosen uniformly at random
// from the subscribers.
//
func (b *SyncTower) Signal(val interface{}) {
	b.mu.Lock()
	n := len(b.names)
	i := rand.Intn(n)
	ch := b.subscribers[b.names[i]]

	// drain first, any old value
	select {
	case <-ch:
	default:
	}
	// then fill with new
	ch <- val
	b.mu.Unlock()
}
