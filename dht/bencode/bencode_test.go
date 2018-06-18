package bencode

import (
	"testing"
	"math"
	"fmt"
)

func isDoubleEqual(f1, f2 float64) bool {
	const TOLERANCE = 0.000001
	return math.Abs(f1-f2) < TOLERANCE
}

func TestDecode(t *testing.T) {

	// 字符串
	packet, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	if packet.Value.GetString() != "spam" {
		t.Fatal("err")
	}

	// 数字
	packet, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	if packet.Value.GetNumber() != 42 {
		t.Fatal("err")
	}

	// 数组
	packet, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	if packet.Value.GetArray()[0].GetString() != "spam" {
		t.Fatal("err")
	}

	if packet.Value.GetArray()[1].GetNumber() != 42 {
		t.Fatal("err")
	}

	// 对象
	packet, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	if packet.Value.GetObject()["bar"].GetString() != "spam" {
		t.Fatal("err")
	}

	if packet.Value.GetObject()["foo"].GetNumber() != 42 {
		t.Fatal("err")
	}
}

func TestEncode(t *testing.T) {

	// 字符串
	packet, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	buf, err := packet.Encode()
	if err != nil || buf != "4:spam" {
		t.Fatal("err", buf)
	}

	// 数字
	packet, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	buf, err = packet.Encode()
	if err != nil || buf != "i42e" {
		t.Fatal("err", buf)
	}

	// 数组
	packet, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf, err = packet.Encode()
	if err != nil || buf != "l4:spami42ee" {
		t.Fatal("err", buf)
	}

	// 对象
	packet, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf, err = packet.Encode()
	if err != nil || buf != "d3:bar4:spam3:fooi42ee" {
		t.Fatal("err", buf)
	}
}

func TestPrettify(t *testing.T) {
	packet, err := Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(packet.Value.Prettify())

}
