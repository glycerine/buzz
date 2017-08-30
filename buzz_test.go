package buzz_test

import (
	"github.com/glycerine/buzz"
	"testing"
)

func TestBuzz(t *testing.T) {

	z := buzz.NewAsyncTower()
	buzzCh := z.Subscribe("me")
	buzzCh2 := z.Subscribe("you")

	select {
	case <-buzzCh:
		t.Fatal("buzzCh starts open; it should have blocked")
	default:
		// ok, good.
	}

	aligator := "bill"
	z.Broadcast(aligator)

	select {
	case b := <-buzzCh:
		if b != aligator {
			t.Fatal("Broadcast(aligator) means aligator should be read on the buzzCh")
		}
	default:
		t.Fatal("buzzCh is now closed, bad. Instead, it should have read back aligator. Refresh should have restocked us.")
	}

	select {
	case b := <-buzzCh2:
		if b != aligator {
			t.Fatal("Broadcast(aligator) means aligator should be read on the buzzCh2")
		}
	default:
		t.Fatal("buzzCh2 is now closed, bad. Instead, it should have read back aligator. Refresh should have restocked us.")
	}

	// multiple Set are fine:
	crocadile := "lyle"
	z.Broadcast(crocadile)

	select {
	case b := <-buzzCh:
		if b != crocadile {
			t.Fatal("Broadcast(crocadile) means crocadile should be read on the buzzCh")
		}
	default:
		t.Fatal("buzzCh is now closed, bad. Instead, it should have read back crocadile. Refresh should have restocked us.")
	}

	select {
	case b := <-buzzCh2:
		if b != crocadile {
			t.Fatal("Broadcast(crocadile) means crocadile should be read on the buzzCh2")
		}
	default:
		t.Fatal("buzzCh2 is now closed, bad. Instead, it should have read back crocadile. Refresh should have restocked us.")
	}

	// and after Clear, we should block
	z.Clear()
	select {
	case <-buzzCh:
		t.Fatal("Clear() means receive should have blocked.")
	case <-buzzCh2:
		t.Fatal("Clear() means receive should have blocked.")
	default:
		// ok, good.
	}

}
