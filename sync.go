package buzz

import (
	"sync"
)

// SyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to SyncTower.
//
// Sends don't block, as subscribers are given buffered channels.
//
type SyncTower struct {
	subs   []chan int
	mut    sync.Mutex
	closed bool

	reqStop chan bool
}

// NewSyncTower makes a new SyncTower.
func NewSyncTower() *SyncTower {
	return &SyncTower{
		reqStop: make(chan bool),
	}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *SyncTower) Subscribe(name string) chan int {
	b.mut.Lock()
	ch := make(chan int, 1)
	b.subs = append(b.subs, ch)
	b.mut.Unlock()
	return ch
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
func (b *SyncTower) Broadcast(val int) {
	for i := range b.subs {
		b.subs[i] <- val // race here, with the close. duh. on purpose.
	}
}

func (b *SyncTower) Close() error {
	b.mut.Lock()
	if b.closed {
		b.mut.Unlock()
		return ErrClosed
	}
	b.closed = true

	for i := range b.subs {
		close(b.subs[i]) // race here, expected.
	}
	b.mut.Unlock()
	return nil
}

func (b *SyncTower) Clear() {
	for i := range b.subs {
		select {
		case <-b.subs[i]:
		default:
		}
	}
}
