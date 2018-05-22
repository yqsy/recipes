package main

import (
	"os"
	"fmt"
	"google.golang.org/grpc"
	pb "github.com/yqsy/recipes/grpc/sudoku/sudoku_protocol"
	"context"
	"time"
)

var usage = `Usage:
%v connectAddr problem
`

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 3 {
		fmt.Printf(usage)
		return
	}

	conn, err := grpc.Dial(arg[1], grpc.WithInsecure())

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	c := pb.NewSudokuSolverClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Solve(ctx, &pb.SolveRequest{Problem: arg[2]})
	if err != nil {
		panic(err)
	}

	if r.Ok {
		fmt.Printf("solve ok: %v\n", r.Result)
	} else {
		fmt.Printf("solve error\n")
	}
}
