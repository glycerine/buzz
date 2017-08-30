package buzz

import (
	//"fmt"
	"sync"
	"testing"
	//"time"
)

func BenchmarkCondToCond(b *testing.B) {
	b.StopTimer()

	cond := sync.NewCond(new(sync.Mutex))

	var val int
	b.SetBytes(int64(8))
	finished := false
	done := make(chan bool)

	go func() {
		// consume odd, produce even integers
		for {
			cond.L.Lock()
			for val%2 == 0 && !finished {
				cond.Wait()
			}
			val++
			cond.Broadcast()
			if finished == true {
				//fmt.Printf("done with consumer loop\n")
				cond.L.Unlock()
				close(done)
				return
			}
			cond.L.Unlock()
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume even, produce odd integers
	for i := 0; i < b.N; i++ {
		cond.L.Lock()
		for val%2 == 1 {
			cond.Wait()
		}
		val++
		cond.Broadcast()
		cond.L.Unlock()
	}
	cond.L.Lock()
	finished = true
	cond.Broadcast()
	cond.L.Unlock()
	b.StopTimer()
	//fmt.Printf("done with producer loop\n")
	<-done
	//fmt.Printf("cond benchmark done!")
}

func BenchmarkAsyncTowerToTower(b *testing.B) {
	b.StopTimer()

	evenTower, oddTower := NewAsyncTower(), NewAsyncTower()
	evenCh := evenTower.Subscribe("even-consumer")
	oddCh := oddTower.Subscribe("odd-consumer")

	var val int
	b.SetBytes(int64(8))
	reqStop, done := make(chan bool), make(chan bool)

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				// send on closed channel
				//fmt.Printf("Broadcast recovered from '%#v'\n", r)
				close(done)
			}
		}()

		// consume odd, produce even integers
		var val2 int
		ok := false
		for {
			val2, ok = <-oddCh // blocked here, done with b.N loop
			if !ok {
				//fmt.Printf("consume odd: odd closed, closing even\n")
				evenTower.Close()
				close(done)
				return
			}
			val2++
			if nil != evenTower.Broadcast(val2) {
				//fmt.Printf("consume odd: even closed, closing odd\n")
				oddTower.Close()
				close(done)
				return
			}
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume even, produce odd integers
	val = 1
	oddTower.sub <- val
	ok := false
	for i := 0; i < b.N; i++ {
		val, ok = <-evenCh
		if !ok {
			//fmt.Printf("consume even: even closed, closing odd\n")
			oddTower.Close()
			break
		}
		val++
		//oddCh <- val
		if nil != oddTower.Broadcast(val) {
			//fmt.Printf("consume even: odd closed, closing even\n")
			evenTower.Close()
			break
		}
	}
	b.StopTimer()
	//fmt.Printf("done with b.N loop\n")
	oddTower.Close()
	evenTower.Close()
	//fmt.Printf("done with b.N loop and towers closed\n")

	// drain
	/*	select {
		case <-evenCh:
		case <-time.After(time.Second):
		}
	*/
	//fmt.Printf("done with producer loop\n")
	close(reqStop)
	<-done
	//	fmt.Printf("tower benchmark done!")
}

func BenchmarkSyncTowerToTower(b *testing.B) {
	b.StopTimer()

	evenTower, oddTower := NewSyncTower(), NewSyncTower()
	evenCh := evenTower.Subscribe("even-consumer")
	oddCh := oddTower.Subscribe("odd-consumer")

	var val int
	b.SetBytes(int64(8))
	reqStop, done := make(chan bool), make(chan bool)

	go func() {
		defer func() {
			r := recover()
			if r != nil {
				// send on closed channel
				//fmt.Printf("Broadcast recovered from '%#v'\n", r)
				close(done)
			}
		}()
		// consume odd, produce even integers
		var val2 int
		ok := false
		for {
			val2, ok = <-oddCh // blocked here, done with b.N loop
			if !ok {
				//fmt.Printf("consume odd: odd closed, closing even\n")
				evenTower.Close()
				close(done)
				return
			}
			val2++
			if nil != evenTower.Broadcast(val2) {
				//fmt.Printf("consume odd: even closed, closing odd\n")
				oddTower.Close()
				close(done)
				return
			}
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume even, produce odd integers
	val = 1
	oddTower.sub <- val
	ok := false
	for i := 0; i < b.N; i++ {
		val, ok = <-evenCh
		if !ok {
			//fmt.Printf("consume even: even closed, closing odd\n")
			oddTower.Close()
			break
		}
		val++
		//oddCh <- val
		if nil != oddTower.Broadcast(val) {
			//fmt.Printf("consume even: odd closed, closing even\n")
			evenTower.Close()
			break
		}
	}
	b.StopTimer()
	//fmt.Printf("done with b.N loop\n")
	oddTower.Close()
	evenTower.Close()
	//fmt.Printf("done with b.N loop and towers closed\n")

	// drain
	/*	select {
		case <-evenCh:
		case <-time.After(time.Second):
		}
	*/
	//fmt.Printf("done with producer loop\n")
	close(reqStop)
	<-done
	//	fmt.Printf("tower benchmark done!")
}

/*
BenchmarkCondToCond-4          	                 3000000	       475 ns/op	  16.83 MB/s
BenchmarkAsyncTowerToTower-4 (actually sync)   	 3000000	       445 ns/op	  17.94 MB/s
BenchmarkSyncTowerToTower-4    	                 2000000	       666 ns/op	  12.00 MB/s

add function wrapper around async send:
BenchmarkCondToCond-4          	 3000000	       464 ns/op	  17.22 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       429 ns/op	  18.61 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       433 ns/op	  18.44 MB/s

BenchmarkCondToCond-4          	 3000000	       477 ns/op	  16.77 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       434 ns/op	  18.39 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       436 ns/op	  18.34 MB/s

with mutex protection of close and function wrapper around send, but no select on Broadcast:
BenchmarkCondToCond-4          	 3000000	       483 ns/op	  16.53 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       441 ns/op	  18.11 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       486 ns/op	  16.43 MB/s

async has select but just one case
BenchmarkCondToCond-4          	 3000000	       479 ns/op	  16.68 MB/s
BenchmarkAsyncTowerToTower-4   	 5000000	       478 ns/op	  16.72 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       484 ns/op	  16.52 MB/s

async has select with 2nd channel in it
BenchmarkCondToCond-4          	 3000000	       473 ns/op	  16.91 MB/s
BenchmarkAsyncTowerToTower-4   	 2000000	       646 ns/op	  12.38 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       447 ns/op	  17.88 MB/s

provacative go blog title: Unintuitive Go: why you should send on closed channels for performance

with the defer recover OUTSIDE of the tight send loop, we actually
get better performance with channels

but sync had the recover on the inner loop on these two
BenchmarkCondToCond-4          	 3000000	       461 ns/op	  17.33 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       415 ns/op	  19.24 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       559 ns/op	  14.29 MB/s

BenchmarkCondToCond-4          	 3000000	       467 ns/op	  17.09 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       422 ns/op	  18.94 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       577 ns/op	  13.86 MB/s

async AND sync outer looped:
BenchmarkCondToCond-4          	 3000000	       455 ns/op	  17.56 MB/s
BenchmarkAsyncTowerToTower-4   	 5000000	       422 ns/op	  18.92 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       424 ns/op	  18.84 MB/s

BenchmarkCondToCond-4          	 3000000	       448 ns/op	  17.85 MB/s
BenchmarkAsyncTowerToTower-4   	 3000000	       425 ns/op	  18.78 MB/s
BenchmarkSyncTowerToTower-4    	 3000000	       419 ns/op	  19.07 MB/s

*/
