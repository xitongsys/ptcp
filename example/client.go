package main

import (
	"fmt"

	"github.com/xitongsys/ptcp/ptcp"
)

func main() {
	ptcp.Init("eth0")
	conn, err := ptcp.Dial("ptcp", "47.240.40.78:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("connected.")

	for i := 0; i < 5; i++ {
		data := conn.LocalAddr().String()
		n, err := conn.Write([]byte("From: " + data))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(n, err)
		fmt.Scanf("%d", &n)
	}
	conn.Close()
}
