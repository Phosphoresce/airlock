package main

import "os"
import "fmt"
import "net"
import "log"

type circle struct {
	peers []*peer
	name string
}

func newCircle(peers []*peer, name string) *circle { return &circle{peers, name} }

type peer struct {
	addr *net.UDPAddr
}

func newPeer(ip string, port int) *peer {
	return &peer{
		addr: &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
	}
}

func main() {
	fmt.Println("Hello, Airlock!")

	p := make([]*peer, 1)
	// peer[0] will always be the local peer
	p[0] = newPeer("127.0.0.1", 9001)
	c := newCircle(p, "local")

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
	}
	listen(c.peers[0])
}

// incoming connections should create a new peer
func listen(local *peer) {
	listener, err := net.ListenUDP("udp", local.addr)

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

// outgoing connections should create a new peer
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
