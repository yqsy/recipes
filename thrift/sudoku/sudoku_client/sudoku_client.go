package main

import (
	"os"
	"fmt"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/yqsy/recipes/thrift/sudoku/sudoku_protocol/gen-go/sudoku_protocol"
	"context"
)

var usage = `Usage:
%v connectAddr problems
`

var defaultCtx = context.Background()

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 3 {
		fmt.Printf(usage)
		return
	}

	var protocolFactory thrift.TProtocolFactory
	var transportFactory thrift.TTransportFactory

	protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory = thrift.NewTTransportFactory()

	var transport thrift.TTransport

	transport, err := thrift.NewTSocket(arg[1])
	if err != nil {
		panic(err)
	}

	transport, err = transportFactory.GetTransport(transport)
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	err = transport.Open()
	if err != nil {
		panic(err)
	}

	iprot := protocolFactory.GetProtocol(transport)
	oprot := protocolFactory.GetProtocol(transport)
	client := sudoku_protocol.NewSudoSolverClient(thrift.NewTStandardClient(iprot, oprot))

	r, err := client.Solve(defaultCtx, &sudoku_protocol.SolveRequest{Problem: arg[2]})
	if err != nil {
		panic(err)
	}

	if r.Ok {
		fmt.Printf("solve ok: %v\n", r.Result_)
	} else {
		fmt.Printf("solve error\n")
	}

}
