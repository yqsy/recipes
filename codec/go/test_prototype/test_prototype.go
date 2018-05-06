package main

import (
	"github.com/yqsy/recipes/codec/go/proto"
	"github.com/golang/protobuf/proto"
	"reflect"
	"fmt"
)

func createMessage(typeName string) (interface{}, error) {
	mt := proto.MessageType(typeName)
	if mt == nil {
		fmt.Errorf("unknown message type %q", typeName)
	}

	return reflect.New(mt.Elem()).Interface(), nil
}

func fooQuery(message proto.Message) {
	query := message.(*codec.Query)
	_ = query
	fmt.Println("query")
}

func fooEmpty(message proto.Message) {
	empty := message.(*codec.Empty)
	_ = empty
	fmt.Println("empty")
}

func main() {
	typeName := proto.MessageName(&codec.Query{})
	fmt.Println(typeName)

	msg, err := createMessage(typeName)

	if err != nil {
		panic(err)
	}

	type Callback func(message proto.Message)
	callbacks := make(map[interface{}]Callback)
	callbacks[reflect.TypeOf((*codec.Query)(nil))] = fooQuery
	callbacks[reflect.TypeOf((*codec.Empty)(nil))] = fooEmpty

	cb := callbacks[reflect.TypeOf(msg)]

	if cb == nil {
		panic("cb err")
	}
	cb(msg.(proto.Message))
}
