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
	"sync"
	//"sync/atomic"
)

// AsyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to AsyncTower.
//
// Sends don't block, as subscribers are given buffered channels.
//
type AsyncTower struct {
	sub    chan int
	mut    sync.Mutex
	closed bool
}

// NewAsyncTower makes a new AsyncTower.
func NewAsyncTower() *AsyncTower {
	return &AsyncTower{}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *AsyncTower) Subscribe(name string) chan int {
	b.mut.Lock()
	ch := make(chan int, 1)
	b.sub = ch
	b.mut.Unlock()
	return ch
}

var ErrClosed = fmt.Errorf("channel closed")

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
func (b *AsyncTower) Broadcast(val int) error {
	b.mut.Lock()
	if !b.closed {
		b.mut.Unlock()
		b.sub <- val
		return nil
	}
	b.mut.Unlock()
	return ErrClosed
}

func (b *AsyncTower) Close() error {
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

func (b *AsyncTower) Clear() {
	select {
	case <-b.sub:
	default:
	}
}

/*

BenchmarkCondToCond-4          	 3000000	       477 ns/op	  16.77 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       434 ns/op	  18.39 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       436 ns/op	  18.34 MB/s

BenchmarkCondToCond-4          	 3000000	       462 ns/op	  17.29 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       451 ns/op	  17.73 MB/s
BenchmarkSyncTowerToTower-4    	 5000000	       433 ns/op	  18.46 MB/s

add don't send on closed protection


*/
