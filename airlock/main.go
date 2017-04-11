package main

import "os"
import "fmt"
import "net"
import "log"
import "time"
import "crypto/sha256"
import "encoding/base64"
import "strings"
import "strconv"
import "bufio"

type circle struct {
	peers []*peer
	name  string
}

func newCircle(peers []*peer, name string) *circle { return &circle{peers, name} }

type peer struct {
	name string
	addr *net.UDPAddr
	time time.Time
	// TODO: save udp client here
	client *net.UDPConn
	userid string
}

func newPeer(ip string, port int) *peer {
	return &peer{
		addr: &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		time: time.Now(),
		userid: func() string {
			id := sha256.Sum256([]byte(time.Now().String()))
			return base64.URLEncoding.EncodeToString(id[:])
		}(),
	}
}

type message struct {
	userid  string
	cmdflag bool
	body    string
}

func main() {
	fmt.Println("Hello, Airlock!")

	// default config values
	var target, username string
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
		case "-u", "--username":
			username = args[x+1]
			x++
		}
	}

	// create structs
	p := make([]*peer, 1)

	// peer[0] will always be the local peer
	p[0] = newPeer("127.0.0.1", port)
	fmt.Printf("userid: %v\n", p[0].userid[:8])
	c := newCircle(p, "local")
	c.peers[0].name = username

	// connecting to peer?
	if target != "" {
		targetAddr := strings.Split(target, ":")
		targetPort, _ := strconv.Atoi(targetAddr[len(targetAddr)-1])
		c.peers = append(c.peers, newPeer(targetAddr[0], targetPort))
		c.connect()
	}

	go c.listen()
	c.chat()
}

// incoming connections should create a new peer and send an updated peer list
func (c *circle) listen() {
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
			message := string(buff[:rlen])
			if strings.Contains(message, ">") {
				fmt.Printf("%v %s\n", remote.IP.String(), message)

				// parse system commands
			} else if strings.HasPrefix(message, "/") {
				c.cmdEngine(remote, message)
			} else {
				fmt.Printf("New connect from %v\n", remote)
				// this response should be a list of ips and ports for all peers BESIDES the currently connected peer and the local peer
				// can avoid appending the connected peer to the list until after we have sent the peer list barring peer[0]
				// build the peer list
				fmt.Printf("Peer objects: %v\n", c.peers)
				if len(c.peers) > 1 {
					peerlist := make([]string, 0)
					for _, peer := range c.peers[1:] {
						peerlist = append(peerlist, peer.addr.String())
					}

					fmt.Printf("sent peerlist: %s\n", strings.Join(peerlist, ","))
					listener.WriteTo([]byte(strings.Join(peerlist, ",")), remote)
				} else {
					listener.WriteTo([]byte("nil"), remote)
				}

				// is remote already a peer? if not add it as a peer
				remotePort, _ := strconv.Atoi(message)

				if peerExists(c, remote.IP.String(), remotePort) == -1 {
					c.peers = append(c.peers, newPeer(remote.IP.String(), remotePort))
				}
			}
		}
	}
}

func (c *circle) cmdEngine(remote *net.UDPAddr, message string) {
	switch {
	case strings.Contains(message, "quit"):
		readbuff := strings.Split(message, ",")
		remotePort, _ := strconv.Atoi(readbuff[len(readbuff)-1])

		i := peerExists(c, remote.IP.String(), remotePort)
		if i != -1 {
			copy(c.peers[i:], c.peers[i+1:])
			c.peers[len(c.peers)-1] = nil
			c.peers = c.peers[:len(c.peers)-1]
		}

	case strings.Contains(message, "heartbeat"):
		// stop the peer from being removed
		fmt.Println("processing heartbeat")
		readbuff := strings.Split(message, ",")
		remotePort, _ := strconv.Atoi(readbuff[len(readbuff)-1])

		j := peerExists(c, remote.IP.String(), remotePort)
		if j != -1 {
			// reset some timeout value
			c.peers[j].time = time.Now()
			fmt.Printf("time set to: %v\n", c.peers[j].time)
		}
	}
}

// outgoing connections should create a new peer from specified target and receive target's peer list
func (c *circle) connect() {
	client, err := net.DialUDP("udp", nil, c.peers[1].addr)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}
	defer client.Close()

	// Send port number this client will listen on.
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

			if peerExists(c, peerAddr[0], peerPort) == -1 {
				c.peers = append(c.peers, newPeer(peerAddr[0], peerPort))
				client2, _ := net.DialUDP("udp", nil, c.peers[len(c.peers)-1].addr)
				defer client2.Close()
				client2.Write(buff)
			}
		}
	}
	go c.heartbeat()
}

// check for peer in peerlist
func peerExists(c *circle, ip string, port int) int {
	fmt.Printf("checking for %v:%v\n", ip, port)
	for i, peer := range c.peers {
		if peer.addr.IP.String() == ip && peer.addr.Port == port {
			return i
		}
	}
	return -1
}

// Send a message to all peers besides self
func (c *circle) chat() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		buffer := scanner.Text()
		for _, peer := range c.peers[1:] {
			client, _ := net.DialUDP("udp", nil, peer.addr)
			if strings.Contains(string(buffer), "/quit") {
				// send quit command instead of text
				port := strconv.Itoa(c.peers[0].addr.Port)
				if !peer.isIdle() {
					// TODO: update to send message struct with userid instead of port
					client.Write([]byte("/quit," + port))
				}
			} else {
				if !peer.isIdle() {
					var user string
					if c.peers[0].name == "" {
						user = c.peers[0].userid
					} else {
						user = c.peers[0].name
					}
					client.Write([]byte(user + " > " + buffer))
				}
			}
			client.Close()
		}
		if strings.Contains(string(buffer), "/quit") {
			return
		}
	}
}

func (p *peer) isIdle() bool {
	// before I send to anybody I need to make sure they are not idle
	// essentially means peer.time - time.Now() < 10 minutes
	return time.Since(p.time) > (10 * time.Minute)
}

func (c *circle) heartbeat() {
	// want to send a control message at a designated interval
	// timer 5 minutes
	for {
		for _, peer := range c.peers[1:] {
			client, _ := net.DialUDP("udp", nil, peer.addr)
			port := strconv.Itoa(c.peers[0].addr.Port)
			client.Write([]byte("/heartbeat," + port))
			client.Close()
		}
		time.Sleep(5 * time.Minute)
	}
}
