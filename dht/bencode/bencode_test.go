package bencode

import (
	"testing"
)

func TestDecode(t *testing.T) {

	// 字符串
	packet, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	if packet.Value.GetString() != "spam" {
		t.Fatal(err)
	}



}
