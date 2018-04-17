package common

import "net"

// 作为服务端要维护多个context,有两种主要的功能1. 类似ssh -NL 2. 类似ssh -NR
type Context struct {

	// 读channel阻塞read
	Channel net.Conn

	

}