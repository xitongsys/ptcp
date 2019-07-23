package main

import (
	"fmt"
	"time"

	"github.com/xitongsys/ptcp/ptcp"
)

func main() {
	ptcp.Init("eth0")
	ln, err := ptcp.Listen("ptcp", "172.19.36.185:12222")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		if conn, err := ln.Accept(); err == nil {
			fmt.Println("new connection: ", conn.RemoteAddr())
			go func() {
				for {
					conn.Write([]byte(fmt.Sprintf("[%v] Hello", time.Now())))
					time.Sleep(time.Second)
				}
			}()

			go func() {
				buf := make([]byte, 100)
				for {
					n, err := conn.Read(buf)
					if err == nil {
						fmt.Printf("From %v: %v\n", conn.RemoteAddr(), string(buf[:n]))
					} else {
						fmt.Printf("%v error: %v\n", conn.RemoteAddr(), err)
						break
					}
				}
			}()
		}
	}

}
