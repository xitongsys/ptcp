package main

import (
	"fmt"

	"github.com/xitongsys/ptcp/ptcp"
)

func main(){
	ptcp.Init("eth0")
	ln, err := ptcp.Listen("ptcp", "127.0.0.1:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 100)
	for {
		n, err := conn.Read(buf)
		fmt.Println(n, err, string(buf[:n]))
	}
}
