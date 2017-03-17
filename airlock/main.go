package main

import "os"
import "fmt"
import "net"
import "log"
import "strconv"
import "bufio"

type circle struct {
	peers []*peer
	name  string
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

	// default config values
	var target string
	port := 9001

	// process args
	args := os.Args[1:]
	for x := 0; x < len(args); x++ {
		switch args[x] {
		case "-t", "--target":
			target = args[x+1]
			x++
		case "-p", "--port":
			port, _ = strconv.Atoi(args[x+1])
			x++
		}
	}

	// create structs
	p := make([]*peer, 1)

	// peer[0] will always be the local peer
	p[0] = newPeer("127.0.0.1", port)
	c := newCircle(p, "local")

	// connecting to peer?
	if target != "" {
		connect(target)
	}
	go listen(c)
	chat(c)
}

// incoming connections should create a new peer
func listen(c *circle) {
	listener, err := net.ListenUDP("udp", c.peers[0].addr)
	if err != nil {
		log.Fatal("Failed to create listener: ", err)
	}
	defer listener.Close()

	for {
		var buff [1024]byte
		rlen, remote, err := listener.ReadFromUDP(buff[:])
		if err != nil {
			log.Print("Failed to read from socket: ", err)
		} else {
			fmt.Printf("received: %v bytes from %v.\nmessage: %s\n", rlen, remote, buff)

			// TODO: is remote already a peer? if not add it as a peer
			c.peers = append(c.peers, &peer{addr: remote})

			response := []byte("Hello Client")
			listener.Write(response)
		}
	}
}

// TODO: Send a message to all peers besides self
func chat(c *circle) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("Send: %v to %v", scanner.Text(), c.peers[1:])
	}
}

// outgoing connections should create a new peer
func connect(target string) {
	client, err := net.Dial("udp", target)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}

	buff := []byte("Hello Server")
	_, err = client.Write(buff)
	if err != nil {
		log.Print("Failed to send: ", err)
	}
}
