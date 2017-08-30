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
	"sync/atomic"
)

// AsyncTower is an 1:M non-blocking value-loadable channel.
//
// Each subscriber gets their own private channel, and it
// will get a copy of whatever is sent to AsyncTower.
//
// Sends don't block, as subscribers are given buffered channels.
//
type AsyncTower struct {
	sub       chan int
	mut       sync.Mutex
	closed    bool
	hasClosed int32

	reqStop chan bool
}

// NewAsyncTower makes a new AsyncTower.
func NewAsyncTower() *AsyncTower {
	return &AsyncTower{
		reqStop: make(chan bool),
	}
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
var ErrShutdown = fmt.Errorf("channel shut down")

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
func (b *AsyncTower) Broadcast(val int) (err error) {
	/*
		defer func() {
			r := recover()
			if r != nil {
				// send on closed channel
				//fmt.Printf("Broadcast recovered from '%#v'\n", r)
				err = ErrClosed
			}
		}()
	*/
	b.sub <- val
	return nil
}

func (b *AsyncTower) Close() error {
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

defer recover
BenchmarkCondToCond-4          	 3000000	       467 ns/op	  17.11 MB/s
Broadcast recovered from '"send on closed channel"'
BenchmarkAsyncTowerToTower-4   	 3000000	       556 ns/op	  14.38 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       483 ns/op	  16.55 MB/s

BenchmarkCondToCond-4          	 3000000	       474 ns/op	  16.86 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       515 ns/op	  15.53 MB/s
BenchmarkSyncTowerToTower-4    	 5000000	       480 ns/op	  16.66 MB/s

defer recover in Async, not in sync
BenchmarkCondToCond-4          	 3000000	       474 ns/op	  16.86 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       515 ns/op	  15.53 MB/s
BenchmarkSyncTowerToTower-4    	 5000000	       480 ns/op	  16.66 MB/s

defer recover in both async and sync
BenchmarkCondToCond-4          	 3000000	       475 ns/op	  16.81 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       573 ns/op	  13.96 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       557 ns/op	  14.36 MB/s

BenchmarkCondToCond-4          	 3000000	       473 ns/op	  16.90 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       572 ns/op	  13.97 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       542 ns/op	  14.74 MB/s

BenchmarkCondToCond-4          	 3000000	       472 ns/op	  16.94 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       560 ns/op	  14.27 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       554 ns/op	  14.43 MB/s


*/
