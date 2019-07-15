package main

import (
	"fmt"

	"github.com/xitongsys/ptcp/ptcp"
)

func main(){
	ptcp.Init("eth0")
	conn, err := ptcp.Dial("ptcp", "47.240.40.78:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("connected.")

	for {
		n, err := conn.Write([]byte("hehe"))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(n, err)
		fmt.Scanf("%d", &n)
	}
}
