package main

import (
	"os"
	"fmt"
	"net"
	"database/sql"
	pb "github.com/yqsy/recipes/recharge/recharge/recharge_protocol"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"github.com/yqsy/recipes/recharge/recharge/service"
	"google.golang.org/grpc/reflection"
)

var usage = `Usage:
%v dbpassword grpcListenAddr
`

func main() {
	arg := os.Args

	usage = fmt.Sprintf(usage, arg[0])

	if len(arg) < 3 {
		fmt.Printf(usage)
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("root:%v@/recharge?charset=utf8", arg[1]))
	if err != nil {
		panic(err)
	}

	defer db.Close()

	listener, err := net.Listen("tcp", arg[2])
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	handler := &service.Handler{DB: db}
	pb.RegisterUserServer(s, handler)
	pb.RegisterRechargeServer(s, handler)
	reflection.Register(s)
	if err := s.Serve(listener); err != nil {
		panic(err)
	}
}
