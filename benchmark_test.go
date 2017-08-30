package buzz

import (
	"sync"
	"testing"
	"time"
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
		// consume odd, produce even integers
		var val2 int
		ok := false
		for {
			val2, ok = <-oddCh
			if !ok {
				close(done)
				return
			}
			val2++
			evenCh <- val2
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume even, produce odd integers
	val = 1
	oddTower.sub <- val
	for i := 0; i < b.N; i++ {
		val = <-evenCh
		val++
		oddCh <- val
	}
	b.StopTimer()
	close(oddCh)

	// drain
	select {
	case <-evenCh:
	case <-time.After(time.Second):
	}
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
		// consume odd, produce even integers
		var val2 int
		ok := false
		for {
			val2, ok = <-oddCh
			if !ok {
				close(done)
				return
			}
			val2++
			evenCh <- val2
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume even, produce odd integers
	val = 1
	oddTower.sub <- val
	for i := 0; i < b.N; i++ {
		val = <-evenCh
		val++
		oddCh <- val
	}
	b.StopTimer()
	close(oddCh)

	// drain
	select {
	case <-evenCh:
	case <-time.After(time.Second):
	}
	//fmt.Printf("done with producer loop\n")
	close(reqStop)
	<-done
	//	fmt.Printf("tower benchmark done!")
}

/*
BenchmarkCondToCond-4          	                 3000000	       475 ns/op	  16.83 MB/s
BenchmarkAsyncTowerToTower-4 (actually sync)   	 3000000	       445 ns/op	  17.94 MB/s
BenchmarkSyncTowerToTower-4    	                 2000000	       666 ns/op	  12.00 MB/s

*/
