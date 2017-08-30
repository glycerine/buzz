// package buzz provides 1:M value-broadcasting channels.
//
// The sending is handled by the main object, the Broadcaster.
// Call buzz.New() to obtain one.
//
// Receivers get their own personal channels.
// To get their channel, a Receivers will call Subscribe()
// on the Broadcaster with
// their string identifier. They obtain a channel
// to receive on. They should never close this channel,
// and should never send on it.
// Receviers can unsubscribe using Unsub().
//
// To send to all receivers, call Broadcaster.Bcast().
//
// Upon broadcast via Bcast(), each subscriber will
// have a copy of the broadcast value in their 1-buffered channel
// to read when they like.
//
// Broadcaster.Clear() will stop broadcasting that value, and empty any
// unconsumed values from each of the individual subscription
// channels.
package buzz

import (
	"sync"
)

// Broadcaster is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to Broadcaster.
type Broadcaster struct {
	subscribers map[string]chan interface{}
	mu          sync.Mutex
	on          bool
	cur         interface{}
}

func New() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string]chan interface{}),
	}
}

func (b *Broadcaster) Subscribe(name string) chan interface{} {
	b.mu.Lock()
	ch := make(chan interface{}, 1)
	b.subscribers[name] = ch
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsub(name string) {
	b.mu.Lock()
	delete(b.subscribers, name)
	b.mu.Unlock()
}

// Get returns the currently set
// broadcast value.
func (b *Broadcaster) Get() interface{} {
	b.mu.Lock()
	r := b.cur
	b.mu.Unlock()
	return r
}

// Bcast is the common case of doing
// both Set() and then On() together
// to start broadcasting a new value.
//
func (b *Broadcaster) Bcast(val interface{}) {
	b.mu.Lock()
	b.cur = val
	b.drain()
	b.on = true
	b.fill()
	b.mu.Unlock()
}

// Clear turns off broadcasting and
// empties the channel of any old values.
func (b *Broadcaster) Clear() {
	b.mu.Lock()
	b.on = false
	b.drain()
	b.cur = nil
	b.mu.Unlock()
}

// drain all messages, leaving b.Ch empty.
// Users typically want Clear() instead.
// Caller must already hold the b.mu.Lock().
func (b *Broadcaster) drain() {
	// empty channels
	for _, ch := range b.subscribers {
		select {
		case <-ch:
		default:
		}
	}
}

// fill up the channels
// Caller must already hold the b.mu.Lock().
func (b *Broadcaster) fill() {
	for _, ch := range b.subscribers {
		select {
		case ch <- b.cur:
		default:
		}
	}
}
