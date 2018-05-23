package main

import (
	"github.com/yqsy/recipes/thrift/sudoku/sudoku_protocol/gen-go/sudoku_protocol"
	"git.apache.org/thrift.git/lib/go/thrift"
	"os"
	"fmt"
	"context"
	"github.com/yqsy/algorithm/sudoku/sudoku_extra"
)

var usage = `Usage:
%v listenAddr
`

type Handler struct{}

func (s *Handler) Solve(ctx context.Context, solveRequest *sudoku_protocol.SolveRequest) (r *sudoku_protocol.SolveReply, err error) {

	table, err := sudoku_extra.ConvertLineToTable(solveRequest.Problem)
	if err != nil {
		return &sudoku_protocol.SolveReply{Ok: false}, nil
	}

	if sudoku_extra.Solve(table, 0) {
		return &sudoku_protocol.SolveReply{Ok: true, Result_: table.GetLine()}, nil
	} else {
		return &sudoku_protocol.SolveReply{Ok: false}, nil
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	var protocolFactory thrift.TProtocolFactory
	var transportFactory thrift.TTransportFactory

	protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory = thrift.NewTTransportFactory()

	transport, err := thrift.NewTServerSocket(arg[1])

	if err != nil {
		panic(err)
	}

	processor := sudoku_protocol.NewSudoSolverProcessor(&Handler{})
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	fmt.Println("Starting the simple server... on ", arg[1])
	err = server.Serve()
	if err != nil {
		panic(err)
	}
}
