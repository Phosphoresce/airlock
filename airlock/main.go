package main

import "os"
import "fmt"
import "net"
import "log"
import "strings"
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
		targetAddr := strings.Split(target, ":")
		targetPort, _ := strconv.Atoi(targetAddr[len(targetAddr)-1])
		c.peers = append(c.peers, newPeer(targetAddr[0], targetPort))
		connect(c)
	}

	go listen(c)
	chat(c)
}

// incoming connections should create a new peer and send an updated peer list
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

			// this response should be a list of ips and ports for all peers BESIDES the currently connected peer and the local peer
			// can avoid appending the connected peer to the list until after we have sent the peer list barring peer[0]
			// build the peer list
			fmt.Println(c.peers)
			if len(c.peers) > 1 {
				peerlist := make([]string, 0)
				for _, peer := range c.peers[1:] {
					peerlist = append(peerlist, peer.addr.String())
				}

				fmt.Printf("peerlist: %s\n", strings.Join(peerlist, ","))
				listener.WriteTo([]byte(strings.Join(peerlist, ",")), remote)
			} else {
				listener.WriteTo([]byte("nil"), remote)
			}

			// TODO: is remote already a peer? if not add it as a peer
			remotePort, _ := strconv.Atoi(string(buff[:rlen]))
			c.peers = append(c.peers, newPeer(remote.IP.String(), remotePort))
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

// outgoing connections should create a new peer from specified target and receive target's peer list
func connect(c *circle) {
	client, err := net.DialUDP("udp", nil, c.peers[1].addr)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}
	defer client.Close()

	// TODO: Send port number this client will listen on.
	buff := []byte(strconv.Itoa(c.peers[0].addr.Port))
	_, err = client.Write(buff)
	if err != nil {
		log.Print("Failed to send: ", err)
	}

	// receive list of peers
	var readbuff [1024]byte
	n, _ := client.Read(readbuff[:])

	// save peers to own circle
	fmt.Printf("received peerlist: '%s' %v bytes\n", string(readbuff[:n]), len(readbuff[:n]))
	if string(readbuff[:n]) != "nil" {
		peerlist := strings.Split(string(readbuff[:n]), ",")
		for _, peer := range peerlist {
			peerAddr := strings.Split(peer, ":")
			peerPort, _ := strconv.Atoi(peerAddr[len(peerAddr)-1])
			c.peers = append(c.peers, newPeer(peerAddr[0], peerPort))
		}
	}
}
