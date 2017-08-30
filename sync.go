package buzz

import (
	"sync"
)

type SyncTower struct {
	sub    chan int
	mut    sync.Mutex
	closed bool
}

// NewSyncTower makes a new SyncTower.
func NewSyncTower() *SyncTower {
	return &SyncTower{}
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
func (b *SyncTower) Broadcast(val int) error {
	b.mut.Lock()
	if !b.closed {
		b.mut.Unlock()
		b.sub <- val
		return nil
	}
	b.mut.Unlock()
	return ErrClosed
}

func (b *SyncTower) Close() error {
	b.mut.Lock()
	if b.closed {
		b.mut.Unlock()
		return ErrClosed
	}
	b.closed = true
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
