# buzz: a 1:N broadcasting channel in Go

## AsyncTower

`buzz.AsyncTower` provides broadcast semantics from
a single source to multiple receivers.

It is a channel compatible replacement for sync.Cond values.

Each sender will receive exactly one copy of the broadcast value. 

A `buzz.AsyncTower` is a re-usable alternative to closing channels.
Closing channels is convenient, but you have to allocate
a new channel every time you want to broadcast. Another
advantage of using `buzz.AsyncTower` is that you get to
convey a value.

Each subscriber is allocated a buffered channel of size 1.

Broadcast and Signal sends are asynchronous, and these methods
never block. They simply put a single value into the buffered
channel of each subscriber.

Upon broadcasting a new value, any old unconsumed values sitting
leftover in the receiver's channel buffers are purged first.

init time:
~~~
// everyone knows about
tower := buzz.NewAsyncTower()
~~~

subscriber side:
~~~
// each subscriber does
ch := tower.Subscribe("me")
select {
  case val := <-ch:
  ...
}

// if necessary, later you can unsubscribe.
tower.Unsub("me")
~~~

publish side:
~~~
// send to all subscribers
tower.Broadcast(val)

// publishers wanting to send to one subscriber at random do
tower.Signal(val)

~~~

## SyncTower

The `buzz.SyncTower`, also in this package, is the same as
`buzz.AsyncTower` except in one respect. `buzz.SyncTower`
uses unbuffered channels. Hence publishers will
block until all subscribers have received the message.

## benchmarks

See the included `benchmark_test.go` file.

~~~
BenchmarkCondToCond-4          	 3000000	       462 ns/op	  17.31 MB/s
BenchmarkAsyncTowerToTower-4   	 1000000	      1103 ns/op	   7.25 MB/s
BenchmarkSyncTowerToTower-4    	 2000000	       889 ns/op	   8.99 MB/s
~~~

In these benchmarks, the channel based Towers are about half the speed of a cond.Sync.

# author

Jason E. Aten, Ph.D.

# license

MIT license
