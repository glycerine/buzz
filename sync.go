package buzz

import (
	"sync"
	"sync/atomic"
)

// SyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to SyncTower.
//
// Sends don't block, as subscribers are given buffered channels.
//
type SyncTower struct {
	sub       chan int
	mut       sync.Mutex
	closed    bool
	hasClosed int32

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
	b.sub = ch
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
func (b *SyncTower) Broadcast(val int) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			// send on closed channel
			//fmt.Printf("Broadcast recovered from '%#v'\n", r)
			err = ErrClosed
		}
	}()
	b.sub <- val
	return nil
}

func (b *SyncTower) Close() error {
	b.mut.Lock()
	if b.closed {
		b.mut.Unlock()
		return ErrClosed
	}
	b.closed = true
	swapped := atomic.CompareAndSwapInt32(&b.hasClosed, 0, 1)
	if !swapped {
		panic("logic error")
	}

	close(b.sub)
	b.mut.Unlock()
	return nil
}

func (b *SyncTower) Clear() {
	select {
	case <-b.sub:
	default:
	}
}
