package main

import (
	"fmt"

	"github.com/xitongsys/ptcp/ptcp"
)

func main(){
	conn, err := ptcp.Dial("ptcp", "127.0.0.1:22222")
	if err != nil {
		fmt.Println(err)
		return
	}

	n, err := conn.Write([]byte("hehe"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(n, err)
}