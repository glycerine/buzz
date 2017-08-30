package buzz

import (
	"fmt"
	"math/rand"
	"sync"
)

// SyncTower is an 1:M blocking channel publisher.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to SyncTower.
//
// Sends always block until all receives have happened.
// Subscribers are given unbuffered channels.
//
type SyncTower struct {
	subs  []chan interface{}
	names []string
	mu    sync.Mutex
}

// NewSyncTower makes a new SyncTower.
func NewSyncTower() *SyncTower {
	return &SyncTower{}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *SyncTower) Subscribe(name string) chan interface{} {
	b.mu.Lock()
	ch := make(chan interface{}, 1)
	b.subs = append(b.subs, ch)
	b.names = append(b.names, name)
	b.mu.Unlock()
	return ch
}

// Unsub will unsubscribe the subscriber named.
func (b *SyncTower) Unsub(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

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
			b.subs = nil
		} else {
			b.names = b.names[1:]
			b.subs = b.subs[1:]
		}

	case k < n-1: // k >= 1 and n >= 2
		b.names = append(b.names[:k], b.names[(k+1):]...)
		b.subs = append(b.subs[:k], b.subs[(k+1):]...)

	case k == n-1:
		b.names = b.names[:(n - 1)]
		b.subs = b.subs[:(n - 1)]
	}
	return nil
}

// Broadcast sends a copy of val to all subs.
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
	for _, ch := range b.subs {
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
// from the subs.
//
func (b *SyncTower) Signal(val interface{}) {
	b.mu.Lock()
	n := len(b.names)
	i := rand.Intn(n)
	ch := b.subs[i]

	// drain first, any old value
	select {
	case <-ch:
	default:
	}
	// then fill with new
	ch <- val
	b.mu.Unlock()
}
