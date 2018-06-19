package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func toBencodeString(in string) string {
	return fmt.Sprintf("%v:%v", len(in), in)
}

// convert to json like string
// use https://jsonlint.com/ to prettify more!
func Prettify(v interface{}) string {
	switch v.(type) {
	case string:
		return `"` + v.(string) + `"`
	case int:
		return strconv.Itoa(v.(int))
	case []interface{}:
		a := v.([]interface{})
		prettify := "["
		for i := 0; i < len(a); i++ {
			prettify += Prettify(a[i]) + ","
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "]"

	case map[string]interface{}:
		o := v.(map[string]interface{})
		prettify := "{"
		for k, v := range o {
			prettify += fmt.Sprintf(`"%v": %v,`, k, Prettify(v))
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "}"
	default:
		panic("only support string,int,[]interface{} and map[string]interface{}")
	}
}

func Encode(v interface{}) string {
	switch v.(type) {
	case string:
		return toBencodeString(v.(string))
	case int:
		return fmt.Sprintf("i%ve", v.(int))
	case []interface{}:
		a := v.([]interface{})
		prettify := "l"
		for i := 0; i < len(a); i++ {
			prettify += Encode(a[i])
		}
		return prettify + "e"
	case map[string]interface{}:
		o := v.(map[string]interface{})
		prettify := "d"
		for k, v := range o {
			prettify += fmt.Sprintf("%v%v", toBencodeString(k), Encode(v))
		}
		return prettify + "e"

	default:
		panic("only support string,int,[]interface{} and map[string]interface{}")
	}
}

func Decode(b string) (interface{}, error) {
	ctx := Context{b: b}
	return ctx.ParseValue()
}

// return obj,remain string,error
func DecodeAndLeak(b string) (interface{}, string, error) {
	ctx := Context{b: b}
	v, err := ctx.ParseValue()
	return v, ctx.b, err
}

type Context struct {
	b string
}

func (ctx *Context) RemoveACharacter(c byte) error {
	if len(ctx.b) < 1 || ctx.b[0] != c {
		return errors.New("syntax error")
	}
	ctx.b = ctx.b[1:]
	return nil
}

func (ctx *Context) PeekACharacter() (byte, error) {
	if len(ctx.b) < 1 {
		return '0', errors.New("syntax error")
	}
	return ctx.b[0], nil
}

func (ctx *Context) GetString() (string, error) {
	p := 0

	for p < len(ctx.b) && isDigit(ctx.b[p]) {
		p++
	}

	Len, err := strconv.ParseInt(ctx.b[:p], 10, 64)
	if err != nil {
		return "", errors.New("syntax error")
	}

	ctx.b = ctx.b[p:]

	if err = ctx.RemoveACharacter(':'); err != nil {
		return "", errors.New("syntax error")
	}

	if len(ctx.b) < int(Len) {
		return "", errors.New("syntax error")
	}

	str := ctx.b[:Len]
	ctx.b = ctx.b[Len:]
	return str, nil
}

func (ctx *Context) ParseString() (string, error) {
	return ctx.GetString()
}

func (ctx *Context) ParseNumber() (int, error) {
	if err := ctx.RemoveACharacter('i'); err != nil {
		return 0, errors.New("syntax error")
	}

	p := 0

	for p < len(ctx.b) && isDigit(ctx.b[p]) {
		p++
	}

	number, err := strconv.ParseInt(ctx.b[:p], 10, 64)
	if err != nil {
		return 0, errors.New("syntax error")
	}

	ctx.b = ctx.b[p:]

	if err = ctx.RemoveACharacter('e'); err != nil {
		return 0, errors.New("syntax error")

	}

	return int(number), nil
}

func (ctx *Context) ParseArray() ([]interface{}, error) {
	if err := ctx.RemoveACharacter('l'); err != nil {
		return nil, errors.New("syntax error")
	}

	a := make([]interface{}, 0)
	for {
		// dispatch
		ele, err := ctx.ParseValue()
		if err != nil {
			return nil, err
		}

		a = append(a, ele)

		// read 'e' represent end of array
		if err := ctx.RemoveACharacter('e'); err == nil {
			break
		}
	}
	return a, nil
}

func (ctx *Context) ParseObject() (map[string]interface{}, error) {
	if err := ctx.RemoveACharacter('d'); err != nil {
		return nil, errors.New("syntax error")
	}

	o := make(map[string]interface{})

	for {
		// read key
		key, err := ctx.GetString()
		if err != nil {
			return nil, err
		}

		// dispatch
		attribute, err := ctx.ParseValue()
		if err != nil {
			return nil, err
		}

		// save to map
		o[key] = attribute

		// read 'e' represent end of object
		if err := ctx.RemoveACharacter('e'); err == nil {
			break
		}
	}

	return o, nil
}

func (ctx *Context) ParseValue() (interface{}, error) {
	c, err := ctx.PeekACharacter()
	if err != nil {
		return nil, err
	}

	if c >= '0' && c <= '9' {
		return ctx.ParseString()
	}

	switch c {
	case 'i':
		return ctx.ParseNumber()
	case 'l':
		return ctx.ParseArray()
	case 'd':
		return ctx.ParseObject()
	default:
		return nil, errors.New(fmt.Sprintf("error character: %v", c))
	}
}
