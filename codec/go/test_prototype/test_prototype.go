package main

import (
	"github.com/yqsy/recipes/codec/go/proto"
	"github.com/golang/protobuf/descriptor"
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

func fooQuery(query *codec.Query) {

}

func fooEmpty(empty *codec.Query) {

}

func main() {
	fd, md := descriptor.ForMessage(&codec.Query{})
	typeName := fd.GetPackage() + "." + md.GetName()

	msg, err := createMessage(typeName)

	if err != nil {
		panic(err)
	}

	fooQuery(msg.(descriptor.Message))

}
