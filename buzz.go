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
	sub chan int
	mu  sync.Mutex
}

// NewAsyncTower makes a new AsyncTower.
func NewAsyncTower() *AsyncTower {
	return &AsyncTower{}
}

// Subscribe returns a new channel that will receive
// all Broadcast values.
func (b *AsyncTower) Subscribe(name string) chan int {
	b.mu.Lock()
	ch := make(chan int, 1)
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
func (b *AsyncTower) Broadcast(val int) {
	b.sub <- val
}

func (b *AsyncTower) Clear() {
	select {
	case <-b.sub:
	default:
	}
}

/*

BenchmarkCondToCond-4          	 3000000	       475 ns/op	  16.83 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       439 ns/op	  18.19 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       449 ns/op	  17.79 MB/s

BenchmarkCondToCond-4          	 3000000	       474 ns/op	  16.86 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       442 ns/op	  18.07 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       447 ns/op	  17.88 MB/s

*/
