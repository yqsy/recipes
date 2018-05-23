package main

import (
	"os"
	"fmt"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/yqsy/recipes/grpc/sudoku/sudoku_protocol"
	"google.golang.org/grpc/reflection"
	"github.com/yqsy/algorithm/sudoku/sudoku_extra"
)

var usage = `Usage:
%v listenAddr
`

type Handler struct{}

func (s *Handler) Solve(ctx context.Context, in *pb.SolveRequest) (*pb.SolveReply, error) {

	table, err := sudoku_extra.ConvertLineToTable(in.Problem)
	if err != nil {
		return &pb.SolveReply{Ok: false}, nil
	}

	if sudoku_extra.Solve(table, 0) {
		return &pb.SolveReply{Ok: true, Result: table.GetLine()}, nil
	} else {
		return &pb.SolveReply{Ok: false}, nil
	}
}

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 2 {
		fmt.Printf(usage)
		return
	}

	listener, err := net.Listen("tcp", arg[1])
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterSudokuSolverServer(s, &Handler{})
	reflection.Register(s)
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
