// package buzz provides 1:M value-broadcasting channels.
//
// A buzz.AsyncTower is a channel-compatible replacement for sync.Cond values.
//
// Call buzz.NewAsyncTower() to start, then each subscriber
// will call Subscribe() on the tower to obtain a channel
// to receive on. They should never close this channel,
// and should never send on it.
//
// Receviers can unsubscribe using Unsub().
//
// To send to all receivers, call AsyncTower.Broadcast().
//
// Upon broadcast via Broadcast(), each subscriber will
// have a copy of the broadcast value in their 1-buffered channel
// to read when they like.
//
// AsyncTower.Clear() will stop broadcasting that value, and empty any
// unconsumed values from each of the individual subscription
// channels.
//
// There is also a SyncTower synchronous version. Upon
// Broadcast, a SyncTower will block until all receivers have
// received the value.
//
package buzz

import (
	"fmt"
	"math/rand"
	"sync"
)

// AsyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to AsyncTower.
//
// Sends don't block, as subscribers are given buffered channels.
//
type AsyncTower struct {
	subs  []chan interface{}
	names []string
	mu    sync.Mutex
}

// NewAsyncTower makes a new AsyncTower.
func NewAsyncTower() *AsyncTower {
	return &AsyncTower{}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *AsyncTower) Subscribe(name string) chan interface{} {
	b.mu.Lock()
	ch := make(chan interface{}, 1)
	b.subs = append(b.subs, ch)
	b.names = append(b.names, name)
	b.mu.Unlock()
	return ch
}

// Unsub will unsubscribe the subscriber named.
func (b *AsyncTower) Unsub(name string) error {
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
// Any old unreceived values are purged
// from the receive queues before sending.
// Since the receivers are all buffered
// channels, Broadcast should never block
// waiting on a receiver.
//
// Any subscriber who subscribes after the Broadcast will not
// receive the Broadcast value, as it is not
// stored internally.
//
func (b *AsyncTower) Broadcast(val interface{}) {
	b.mu.Lock()
	b.drain()
	b.fill(val)
	b.mu.Unlock()
}

// Signal works like sync.Cond's Signal.
// It sends val to exactly one listener.
//
// The listener is chosen uniformly at random
// from the subs.
//
// Any old value leftover in the chosen
// receiver's buffer is purged first.
//
func (b *AsyncTower) Signal(val interface{}) {
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

// Clear empties the channel of any old values.
func (b *AsyncTower) Clear() {
	b.mu.Lock()
	b.drain()
	b.mu.Unlock()
}

// drain all messages, leaving b.Ch empty.
// Users typically want Clear() instead.
// Caller must already hold the b.mu.Lock().
func (b *AsyncTower) drain() {
	// empty channels
	for _, ch := range b.subs {
		select {
		case <-ch:
		default:
		}
	}
}

// fill up the channels
// Caller must already hold the b.mu.Lock().
func (b *AsyncTower) fill(val interface{}) {
	for _, ch := range b.subs {
		select {
		case ch <- val:
		default:
		}
	}
}
