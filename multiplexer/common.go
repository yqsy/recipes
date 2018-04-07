package main

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type PacketHeader struct {
	Len uint32
	Id  uint32
	Cmd bool

	// body:
	// if Cmd == true
	// command string , end with \r\n
	// else
	// payload
}

// command simple:
// SYN to ip(domain):port\r\n
// FIN\r\n

