package main

import (
	"flag"
	"golang.org/x/net/context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"net/http"
	"github.com/golang/glog"
	gw "github.com/yqsy/recipes/recharge/recharge/recharge_protocol"
)

var (
	endpoint = flag.String("endpoint", "localhost:9090", "endpoint of YourService")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := gw.RegisterUserHandlerFromEndpoint(ctx, mux, *endpoint, opts); err != nil {
		return err
	}

	if err := gw.RegisterRechargeHandlerFromEndpoint(ctx, mux, *endpoint, opts); err != nil {
		return err
	}

	return http.ListenAndServe(":8080", mux)
}

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
