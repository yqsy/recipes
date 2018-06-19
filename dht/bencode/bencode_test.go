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

	if value.(string) != "spam" {
		t.Fatal("err")
	}

	// 数字
	value, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	if value.(int) != 42 {
		t.Fatal("err")
	}

	// 数组
	value, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	if value.([]interface{})[0].(string) != "spam" {
		t.Fatal("err")
	}

	if value.([]interface{})[1].(int) != 42 {
		t.Fatal("err")
	}

	// 对象
	value, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	if value.(map[string]interface{})["bar"].(string) != "spam" {
		t.Fatal("err")
	}

	if value.(map[string]interface{})["foo"].(int) != 42 {
		t.Fatal("err")
	}
}

func TestEncode(t *testing.T) {

	// 字符串
	value, err := Decode("4:spam")
	if err != nil {
		t.Fatal(err)
	}

	buf := Encode(value)
	if err != nil || buf != "4:spam" {
		t.Fatal("err", buf)
	}

	// 数字
	value, err = Decode("i42e")
	if err != nil {
		t.Fatal(err)
	}

	buf = Encode(value)
	if err != nil || buf != "i42e" {
		t.Fatal("err", buf)
	}

	// 数组
	value, err = Decode("l4:spami42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf = Encode(value)
	if err != nil || buf != "l4:spami42ee" {
		t.Fatal("err", buf)
	}

	// 对象
	value, err = Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	buf = Encode(value)
	if err != nil || buf != "d3:bar4:spam3:fooi42ee" {
		t.Fatal("err", buf)
	}
}

func TestPrettify(t *testing.T) {
	value, err := Decode("d3:bar4:spam3:fooi42ee")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(Prettify(value))

}

// 常规使用方法
func TestSimpleUse(t *testing.T) {
	// 编码
	fuck := map[string]interface{}{
		"t": 123,
		"y": "q",
		"q": "find_node",
		"a": map[string]interface{}{
			"id":     "12345678901234567890",
			"target": "09876543210987654321",
		},
	}

	buf := Encode(fuck)

	// 解码
	shit, err := Decode(buf)
	if err != nil {
		panic(err)
	}

	// 增加可读性
	fmt.Println(Prettify(shit))

	// 使用
	// 使用时一定要判断 是否是该类型, object要判断是否有key
	if obj, ok := shit.(map[string]interface{}); ok {
		if T, ok := obj["t"]; ok {
			// 必须要做类型检查
			if TT, ok := T.(int); ok {
				if TT != 123 {
					t.Fatal("err")
				}
			}
		}

		if Y, ok := obj["y"]; ok {
			if _, ok := Y.(int); ok {
				t.Fatal("err")
			}
		}
	}

	// 最好是做一层检查层,以便达到IDL的鲁棒性

}

// 待解析字符串还有剩余的部分另做他用
func TestSpecial(t *testing.T) {
	value, remain, err := DecodeAndLeak("d3:bar4:spam3:fooi42ee123abc")
	if err != nil {
		t.Fatal(err)
	}

	if remain != "123abc" {
		t.Fatal("err")
	}

	if value.(map[string]interface{})["bar"].(string) != "spam" {
		t.Fatal("err")
	}

	if value.(map[string]interface{})["foo"].(int) != 42 {
		t.Fatal("err")
	}
}
