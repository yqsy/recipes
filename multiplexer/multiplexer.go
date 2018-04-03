package main

import (
	"os"
	"fmt"
	"strings"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type Config struct {
	RemoteAddr string
	ConnPair   map[string]string // [bind addr]remote connect addr
}

func parseConfig(arg []string) *Config {
	if len(arg) < 3 {
		return nil
	}

	config := new(Config)
	config.RemoteAddr = arg[2]

	config.ConnPair = make(map[string]string)

	pairs := strings.Split(arg[1], ";")
	if len(pairs) < 1 {
		return nil
	}
	for _, pair := range pairs {
		fourPart := strings.Split(pair, ":")

		if len(fourPart) != 4 {
			return nil
		}
		bindAddr := fourPart[0] + ":" + fourPart[1]
		remoteConnectAddr := fourPart[2] + ":" + fourPart[3]

		config.ConnPair[bindAddr] = remoteConnectAddr
	}

	return config
}

func printUsage(exec string) {
	fmt.Printf("Usage:\n"+
		"%v [bind_address]:port:host:hostport;[...] remotehost:remoteport\n"+
		"Example:\n"+
		"%v :5001:localhost:5001 pi1:30000\n", exec, exec)
}

func main() {
	arg := os.Args

	if len(arg) < 3 {
		printUsage(arg[0])
		return
	}

	config := parseConfig(arg)

	if config == nil {
		printUsage(arg[0])
		return
	}

	//listener, err := net.Listen("tcp", arg[0])
	//
	//panicOnError(err)
	//
	//defer listener.Close()
	//
	//for {
	//	_, err := listener.Accept()
	//	if err != nil {
	//		continue
	//	}
	//
	//}

}
