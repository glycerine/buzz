package buzz

import (
	"sync"
)

type SyncTower struct {
	sub chan int
	mu  sync.Mutex
}

// NewSyncTower makes a new SyncTower.
func NewSyncTower() *SyncTower {
	return &SyncTower{}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *SyncTower) Subscribe(name string) chan int {
	b.mu.Lock()
	ch := make(chan int)
	b.sub = ch
	b.mu.Unlock()
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
	b.sub <- val
}

func (b *SyncTower) Clear() {
	select {
	case <-b.sub:
	default:
	}
}
