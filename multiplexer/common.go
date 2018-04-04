package main

type PacketHeader struct {
	Len int
	Id  int
	Cmd byte

	// body:
	// if Cmd == true
	// command string , end with \r\n
	// else
	// payload
}

// command simple:
// SYN to ip(domain):port\r\n
// FIN\r\n

