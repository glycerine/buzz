# buzz: a 1:N broadcasting channel in Go

`buzz.Tower` provides broadcast semantics from
a single source to multiple receivers.

It is a channel compatible replacement for sync.Cond values.

Each sender will receive exactly one copy of the broadcast value. 

A buzz.Tower is a re-usable alternative to closing channels.
Closing channels is convenient, but you have to allocate
a new channel every time you want to broadcast.

~~~
// global setup
tower := buzz.NewTower()
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
// publisher(s) do
tower.Broadcast(val)

// publishers wanting to send to one subscriber at random do
tower.Signal(val)

~~~
