package bencode

import (
	"testing"
	"fmt"
)

func TestDecode(t *testing.T) {

	// 字符串
	value, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	if value.GetString() != "spam" {
		t.Fatal("err")
	}

	// 数字
	value, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	if value.GetNumber() != 42 {
		t.Fatal("err")
	}

	// 数组
	value, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	if value.GetArray()[0].GetString() != "spam" {
		t.Fatal("err")
	}

	if value.GetArray()[1].GetNumber() != 42 {
		t.Fatal("err")
	}

	// 对象
	value, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	if value.GetObject()["bar"].GetString() != "spam" {
		t.Fatal("err")
	}

	if value.GetObject()["foo"].GetNumber() != 42 {
		t.Fatal("err")
	}
}

func TestEncode(t *testing.T) {

	// 字符串
	value, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	buf := value.Encode()
	if err != nil || buf != "4:spam" {
		t.Fatal("err", buf)
	}

	// 数字
	value, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	buf = value.Encode()
	if err != nil || buf != "i42e" {
		t.Fatal("err", buf)
	}

	// 数组
	value, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf = value.Encode()
	if err != nil || buf != "l4:spami42ee" {
		t.Fatal("err", buf)
	}

	// 对象
	value, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf = value.Encode()
	if err != nil || buf != "d3:bar4:spam3:fooi42ee" {
		t.Fatal("err", buf)
	}
}

func TestPrettify(t *testing.T) {
	value, err := Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(value.Prettify())

}

func TestEncodeNew(t *testing.T) {
	// 1. 数组包数组

	var shit = []int{1, 2, 3}

	value, err := NewArray(shit)

}
