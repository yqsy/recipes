package bencode

import (
	"errors"
	"fmt"
	"strconv"
)

type Kind int

const (
	String Kind = 0
	Number      = 1
	Array       = 2
	Object      = 3
)

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'

}

func toBencodeString(in string) string {
	return fmt.Sprintf("%v:%v", len(in), in)
}

type Value struct {
	Kind Kind

	String_ *string

	Number *float64

	Array *[]*Value

	Object *map[string]*Value
}

// convert to json like string
// use https://jsonlint.com/ to prettify more!
func (value *Value) Prettify() string {
	if value.Kind == String {
		return value.GetString()
	} else if value.Kind == Number {
		return fmt.Sprintf("%v", value.GetNumber())
	} else if value.Kind == Array {
		prettify := "["
		a := *value.GetArray()
		for i := 0; i < len(a); i++ {
			prettify += a[i].Prettify() + ","
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "]"
	} else if value.Kind == Object {
		prettify := "{"
		o := *value.GetObject()
		for k, v := range o {
			prettify += fmt.Sprintf("%v: %v,", k, v.Prettify())
		}
		if len(prettify) > 0 && prettify[len(prettify)-1] == ',' {
			prettify = prettify[:len(prettify)-1]
		}
		return prettify + "}"
	} else {
		panic("impossible")
	}
}

func (value *Value) Encode() string {
	if value.Kind == String {
		return toBencodeString(value.GetString())
	} else if value.Kind == Number {
		return fmt.Sprintf("i%ve", value.GetNumber())
	} else if value.Kind == Array {
		prettify := "l"
		a := *value.GetArray()
		for i := 0; i < len(a); i++ {
			prettify += a[i].Prettify()
		}
		return prettify + "e"
	} else if value.Kind == Object {
		prettify := "d"
		o := *value.GetObject()
		for k, v := range o {
			fmt.Sprintf("%v%v", toBencodeString(k), v.Encode())
		}
		return prettify + "e"
	} else {
		panic("impossible")
	}
}

func (value *Value) GetString() string {
	if value.Kind == String || value.String_ != nil {
		return *value.String_
	} else {
		return ""
	}
}

func (value *Value) GetNumber() float64 {
	if value.Kind == Number {
		return *value.Number
	} else {
		return 0.0
	}
}

func (value *Value) GetArray() *[]*Value {
	return value.Array
}

func (value *Value) GetObject() *map[string]*Value {
	return value.Object
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

func (ctx *Context) ParseString() (*Value, error) {
	if string_, err := ctx.GetString(); err != nil {
		return nil, err
	} else {
		return &Value{Kind: String, String_: &string_}, nil
	}
}

func (ctx *Context) ParseNumber() (*Value, error) {
	if err := ctx.RemoveACharacter('i'); err != nil {
		return nil, errors.New("syntax error")
	}

	p := 0

	for p < len(ctx.b) && isDigit(ctx.b[p]) {
		p++
	}

	number, err := strconv.ParseFloat(ctx.b[:p], 64)
	if err != nil {
		return nil, errors.New("syntax error")
	}

	ctx.b = ctx.b[p:]

	if err = ctx.RemoveACharacter('e'); err != nil {
		return nil, errors.New("syntax error")

	}

	return &Value{Kind: Number, Number: &number}, nil
}

func (ctx *Context) ParseArray() (*Value, error) {
	if err := ctx.RemoveACharacter('l'); err != nil {
		return nil, errors.New("syntax error")
	}

	value := &Value{Kind: Array}
	for {
		// dispatch
		ele, err := ctx.ParseValue()
		if err != nil {
			return nil, err
		}

		// save to array
		if value.Array == nil {
			a := make([]*Value, 0)
			value.Array = &a
		}

		*value.Array = append(*value.Array, ele)

		// read 'e' represent end of array
		if err := ctx.RemoveACharacter('e'); err != nil {
			break
		}
	}
	return value, nil
}

func (ctx *Context) ParseObject() (*Value, error) {
	if err := ctx.RemoveACharacter('d'); err != nil {
		return nil, errors.New("syntax error")
	}

	value := &Value{Kind: Object}

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
		if value.Object == nil {
			m := make(map[string]*Value)
			value.Object = &m
		}

		(*value.Object)[key] = attribute

		// read 'e' represent end of object
		if err := ctx.RemoveACharacter('e'); err != nil {
			break
		}
	}

	return value, nil
}

func (ctx *Context) ParseValue() (*Value, error) {
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

type Packet struct {
	value *Value
}

func (p *Packet) Encode() (string, error) {
	if p.value == nil {
		return "", errors.New(" value is nil")
	}
	return p.value.Encode(), nil
}

func Decode(b string) (*Packet, error) {
	ctx := &Context{b: b}
	if value, err := ctx.ParseValue(); err != nil {
		return nil, err
	} else {
		return &Packet{value: value}, nil
	}
}
