package main

import (
	"net"
	"log"
	"bufio"
	"fmt"
)

func main() {
	// client example
	var message string
	conn, err := net.Dial("tcp", "localhost:80")
	if err != nil {
		log.Fatal(err)
	}
	for {
		fmt.Scan(&message)
		fmt.Fprintf(conn, "%v\n", message)
		reply, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Server: ", reply)
	}
}
