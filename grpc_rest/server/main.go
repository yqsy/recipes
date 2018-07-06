package main

import (
	"os"
	"fmt"
	"net"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	pb "github.com/yqsy/recipes/grpc_rest/service_protocol"
	"google.golang.org/grpc/reflection"
)

var usage = `Usage:
%v listenAddr
`

type Handler struct{}

func (s *Handler) Echo(ctx context.Context, in *pb.StringMessage) (*pb.StringMessage, error) {
	fmt.Printf("value: %v\n", in.Value)
	return &pb.StringMessage{Value: in.Value}, nil
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
	pb.RegisterServiceAServer(s, &Handler{})
	reflection.Register(s)
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
