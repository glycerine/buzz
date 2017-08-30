package buzz_test

import (
	"fmt"
	"github.com/glycerine/buzz"
	"time"
)

func Example() {

	b := buzz.New()
	ch := b.Subscribe("me")
	go func() {
		for {
			select {
			case v := <-ch:
				fmt.Printf("received on Ch: %v\n", v)
			}
		}
	}()

	b.Bcast(4)
	time.Sleep(20 * time.Millisecond)
	b.Bcast(5)
	time.Sleep(20 * time.Millisecond)
	b.Clear()
	time.Sleep(20 * time.Millisecond)
	// Output:
	//received on Ch: 4
	//received on Ch: 5
}