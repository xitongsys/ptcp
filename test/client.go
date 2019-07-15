package main

import (
	"fmt"

	"github.com/xitongsys/ptcp/ptcp"
)

func main(){
	ptcp.Init("eth0")
	conn, err := ptcp.Dial("ptcp", "10.172.117.13:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		n, err := conn.Write([]byte("hehe"))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(n, err)
		fmt.Scanf("%d", &n)
	}
}
