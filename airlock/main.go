package main

import "os"
import "fmt"
import "net"
import "log"

func main() {
	fmt.Println("Hello, Airlock!")
	initialize()
}

func initialize() {
	// process args
	var target string = ""
	args := os.Args[1:]
	for x := 0; x < len(args); x++ {
		switch args[x] {
		case "-t", "--target":
			target = args[x+1]
			x++
		}
	}
	// connecting to peer or creating circle?
	if target != "" {
		connect(target)
	} else {
		listen()
	}
	// if connecting to peer dial ip / dns name

	// if starting new circle host listen on specified port
}

func listen() {
	addr := net.UDPAddr{
		Port: 9001,
		IP:   net.ParseIP("127.0.0.1"), // * to listen from anywhere
	}
	listener, err := net.ListenUDP("udp", &addr)

	if err != nil {
		log.Fatal("Failed to create listener: ", err)
	}
	defer listener.Close()

	var buff [1024]byte
	rlen, remote, err := listener.ReadFromUDP(buff[:])
	if err != nil {
		log.Print("Failed to read from socket: ", err)
	}
	fmt.Printf("received: %v bytes from %v.", rlen, remote)

	var response []byte = []byte("Hello Client")
	listener.Write(response)
}

func connect(target string) {
	client, err :=  net.Dial("udp", target)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}

	var buff []byte = []byte("Hello Server")
	_, err = client.Write(buff)
	if err != nil {
		log.Print("Failed to send: ", err)
	}
}
