package buzz_test

import (
	"fmt"
	"github.com/glycerine/buzz"
	"time"
)

func Example() {

	b := buzz.NewAsyncTower()
	ch := b.Subscribe()
	go func() {
		for {
			select {
			case v := <-ch:
				fmt.Printf("received on Ch: %v\n", v)
			}
		}
	}()

	b.Broadcast(4)

	// since we're async, if we don't pause, the
	// subscriber above might miss the 4.
	time.Sleep(20 * time.Millisecond)
	b.Broadcast(5)
	time.Sleep(20 * time.Millisecond)
	b.Clear()
	time.Sleep(20 * time.Millisecond)
	// Output:
	//received on Ch: 4
	//received on Ch: 5
}
