package transaction

import "testing"

func TestSimple(t *testing.T) {
	tm := Transaction{}

	cur := tm.FetchAndAdd()

	if cur != 0 {
		t.Fatalf("err %v", cur)
	}

}

func TestSimple2(t *testing.T) {

	tm := Transaction{}
	for i := 0; i < 100000; i++ {
		if tm.FetchAndAdd() != uint64(i%65536) {
			t.Fatal("err")
		}
	}
}
