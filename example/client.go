package main

import (
	"fmt"
	"time"

	"github.com/xitongsys/ptcp/ptcp"
)

func main() {
	ptcp.Init("eth0", 2)
	conn, err := ptcp.Dial("ptcp", "127.0.0.1:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("connected")
	go func() {
		for {
			buf := make([]byte, 1024)
			if n, err := conn.Read(buf); err == nil {
				fmt.Println("From server: ", string(buf[:n]))
				continue
			}
			break
		}
	}()

	for i := 0; i < 5; i++ {
		_, err := conn.Write([]byte("hello"))
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second)
	}

	if err = conn.Close(); err != nil {
		fmt.Println("close error", err)
	}
}
