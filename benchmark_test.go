package buzz

import (
	"sync"
	"testing"
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

	// produce odd integers
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
		var ival interface{}
		var val2 int
		for {
			select {
			case <-reqStop:
				close(done)
				//fmt.Printf("done with consumer loop\n")
				return
			case ival = <-oddCh:
				val2 = ival.(int)
				val2++
				evenTower.Broadcast(val2)
			}
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume evens, produce odd integers
	val = 1
	oddTower.Broadcast(val)
	for i := 0; i < b.N; i++ {
		select {
		case next := <-evenCh:
			val = next.(int)
			val++
			oddTower.Broadcast(val)
		}
	}
	b.StopTimer()
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
		var ival interface{}
		var val2 int
		for {
			select {
			case <-reqStop:
				close(done)
				//fmt.Printf("done with consumer loop\n")
				return
			case ival = <-oddCh:
				val2 = ival.(int)
				val2++
				evenTower.Broadcast(val2)
			}
		}

	}()

	b.ResetTimer()
	b.StartTimer()

	// consume evens, produce odd integers
	val = 1
	oddTower.Broadcast(val)
	for i := 0; i < b.N; i++ {
		select {
		case next := <-evenCh:
			val = next.(int)
			val++
			oddTower.Broadcast(val)
		}
	}
	b.StopTimer()
	//fmt.Printf("done with producer loop\n")
	close(reqStop)
	<-done
	//	fmt.Printf("tower benchmark done!")
}
