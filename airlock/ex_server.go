package main

import (
	"log"
	"net"
	"bufio"
	"fmt"
)

func main() {
	listen()
}

func listen() {
	// Listens for tcp connections from clients
	listener, err := net.Listen("tcp", ":80")
	if err != nil {
		// handle error?
		log.Fatal("Create Listener: ", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			// handle error
			log.Fatal("Accept Connection: ", err)
		}
		// Handle the conn in new goroutine.
		go buffer(conn)
	}
}

func buffer(conn net.Conn) {
	// Accepts chat from incoming chat client connection
	for {
		message, _ := bufio.NewReader(conn).ReadString('\n')
		if message != "" {
			fmt.Println("Recieved: ", string(message))
			broadcast(message, conn)
			message = ""
		}
	}
}

func broadcast(buffer string, conn net.Conn) {
	// Transmits buffered chat to clients
	conn.Write([]byte(buffer))
}
