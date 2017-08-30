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
	subs      []chan int
	one       chan int
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
	b.subs = append(b.subs, ch)
	b.one = ch
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
func (b *AsyncTower) Broadcast(val int) {
	for i := range b.subs {
		b.subs[i] <- val // race here, with the close. duh. on purpose.
		/* 5% faster
							BenchmarkCondToCond-4          	 3000000	       473 ns/op	  16.91 MB/s
							BenchmarkAsyncTowerToTower-4   	 3000000	       412 ns/op	  19.40 MB/s
							BenchmarkSyncTowerToTower-4    	 3000000	       418 ns/op	  19.12 MB/s

				BenchmarkCondToCond-4          	 3000000	       459 ns/op	  17.41 MB/s
				BenchmarkAsyncTowerToTower-4   	 3000000	       439 ns/op	  18.18 MB/s
				BenchmarkSyncTowerToTower-4    	 3000000	       424 ns/op	  18.85 MB/s

		BenchmarkCondToCond-4          	 3000000	       456 ns/op	  17.54 MB/s
		BenchmarkAsyncTowerToTower-4   	 3000000	       423 ns/op	  18.88 MB/s
		BenchmarkSyncTowerToTower-4    	 3000000	       421 ns/op	  18.97 MB/s
		*/
	}
}

func (b *AsyncTower) Close() error {
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

func (b *AsyncTower) Clear() {
	for i := range b.subs {
		select {
		case <-b.subs[i]:
		default:
		}
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

BenchmarkCondToCond-4          	 3000000	       468 ns/op	  17.09 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       413 ns/op	  19.36 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       425 ns/op	  18.79 MB/s


*/
