package main

import "os"
import "fmt"
import "net"
import "log"
import "time"
import "bytes"
import "crypto/sha256"
import "encoding/base64"
import "encoding/binary"
import "strings"
import "strconv"
import "bufio"

type circle struct {
	peers []*peer
	msgs  []*message
	name  string
}

func newCircle(peers []*peer, msgs []*message, name string) *circle { return &circle{peers, msgs, name} }

type peer struct {
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

func addPeer(id string, ip string, port int) *peer {
	return &peer{
		addr: &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		time:   time.Now(),
		userid: id,
	}
}

type message struct {
	userid  string
	cmdflag string
	body    string
}

// parse messages from recieved byte arrays
func parseMessage(buff []byte) *message {
	msg := &message{}
	reader := bytes.NewReader(buff)
	binary.Read(reader, binary.BigEndian, &msg)
	return msg
}

// TODO: create a method to create messages to send
func packMessage(id string, flag bool, buffer string) *message {
	return &message{
		userid:  id,
		cmdflag: strconv.FormatBool(flag),
		body:    buffer,
	}
}

const delimiter = "\x1f"

func main() {
	fmt.Println("Hello, Airlock!")

	// default config values
	var target string
	username := ""
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
	m := make([]*message, 0)

	// peer[0] will always be the local peer
	p[0] = newPeer("127.0.0.1", port)
	fmt.Printf("userid: %v\n", p[0].userid[:])

	// create local circle
	c := newCircle(p, m, "local")
	if username != "" {
		c.peers[0].userid = username
	}

	// connecting to peer?
	if target != "" {
		// split the ip and port from target arg
		targetAddr := strings.Split(target, ":")
		targetPort, _ := strconv.Atoi(targetAddr[len(targetAddr)-1])

		// ip is the first string: targetAddr[0]
		c.peers = append(c.peers, newPeer(targetAddr[0], targetPort))
		c.connect()
	}

	// execute main listener and chat function
	go c.listen()
	c.chat()
}

// incoming connections should create a new peer and send an updated peer list
func (c *circle) listen() {
	// set up udp listener
	listener, err := net.ListenUDP("udp", c.peers[0].addr)
	if err != nil {
		log.Fatal("Failed to create listener: ", err)
	}
	defer listener.Close()

	// continue listening until the application is closed
	for {
		// rezero the buffer each loop
		var buff [1024]byte

		// read from the UDP listener
		rlen, remote, err := listener.ReadFromUDP(buff[:])

		// check for failure, if yes, continue listening, if retrieved message process it
		if err != nil {
			log.Print("Failed to read from socket: ", err)
		} else {
			// create struct with string from []byte buffer with length specified by the length of data sent to the udp listener
			msg := parseMessage(buff[:rlen])
			fmt.Printf("received buffer: %v\n", string(buff[:rlen]))
			fmt.Printf("received length: %v\n", rlen)
			fmt.Printf("received msg struct: %v\n", msg)
			fmt.Printf("received msg: %s %s %v\n", msg.userid, msg.body, msg.cmdflag)

			// decide if the message is a chat message, or command to be processed by the system
			flag, _ := strconv.ParseBool(msg.cmdflag)
			if !flag {
				// simply print chat messages
				c.msgs = append(c.msgs, msg)
				fmt.Printf("%s > %s\n", msg.userid[:8], msg.body)
				fmt.Printf("%v\n", len(c.msgs))

			} else if flag {
				// if command, send it to the command engine to handle
				c.cmdEngine(msg, remote, listener)
			}
		}
	}
}

func (c *circle) cmdEngine(msg *message, remote *net.UDPAddr, listener *net.UDPConn) {
	switch {
	case strings.Contains(msg.body, "quit"):
		// quit
		fmt.Println("peer quitting...")

		// remove peer if it exists
		i := c.peerExists(msg.userid)
		if i != -1 {
			fmt.Println("removing peer")
			copy(c.peers[i:], c.peers[i+1:])
			c.peers[len(c.peers)-1] = nil
			c.peers = c.peers[:len(c.peers)-1]
		}

	case strings.Contains(msg.body, "heartbeat"):
		// stop the peer from being removed
		fmt.Println("processing heartbeat...")

		// reset peer timeout if exists
		j := c.peerExists(msg.userid)
		if j != -1 {
			c.peers[j].time = time.Now()
			fmt.Printf("time set to: %v\n", c.peers[j].time)
		}
	default:
		// process new connect
		fmt.Printf("new connect from %v\n", remote)

		// this is a new connect, check to see if we need to send a peerlist
		if len(c.peers) > 1 {
			// build the peer list to send
			peerlist := make([]string, 0)

			// grab all peers besides the local peer
			for _, peer := range c.peers[1:] {
				// TODO: need to add the userid AND the address when sending to new connect
				peerlist = append(peerlist, peer.addr.String())
			}

			// this response should be a list of ips and ports for all peers BESIDES the currently connected peer and the local peer
			listener.WriteTo([]byte(c.peers[0].userid+delimiter+strconv.FormatBool(true)+delimiter+strings.Join(peerlist, ",")), remote)
			fmt.Printf("sent peerlist: %s\n", strings.Join(peerlist, ","))
		} else {
			// no peers besides the two talking to each other now, send nil peerlist
			listener.WriteTo([]byte(c.peers[0].userid+delimiter+strconv.FormatBool(true)+delimiter+"nil"), remote)
		}

		// parse out the port from message... this will get removed with standard struct
		remotePort, _ := strconv.Atoi(msg.body)

		// check if the peer exists before adding it to the circle
		index := c.peerExists(msg.userid)
		if index == -1 {
			c.peers = append(c.peers, addPeer(msg.userid, remote.IP.String(), remotePort))
		} else {
			fmt.Println("sending offline messages")
			// convert c.msgs in strings
			// BUG: this executes but the receiving end doesn't get anything. is the client correct?? is the message correct? should probably print these out to make sure
			client, _ := net.DialUDP("udp", nil, c.peers[index].addr)
			for _, item := range c.msgs {
				// send them missed messages
				//client.Write([]byte(item.userid + delimiter + item.cmdflag + delimiter + item.body))
				// BUG: this will not preserve the names of the original sender, should send whole message struct instead of body
				// TODO: this will also check if the peer exists twice? is that an issue?.. probably want to have a move this to a different command
				c.clientWrite(client, c.peers[index].userid, item.body, false)
			}
			client.Close()
		}
	}
}

// outgoing connections should create a new peer from specified target and receive target's peer list
func (c *circle) connect() {
	// create udp client
	client, err := net.DialUDP("udp", nil, c.peers[1].addr)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}
	defer client.Close()

	// port number this client will listen on.
	buff := strconv.Itoa(c.peers[0].addr.Port)

	// send message struct format, buff is port
	//client.Write([]byte(c.peers[0].userid + delimiter + strconv.FormatBool(true) + delimiter + buff))
	c.clientWrite(client, c.peers[1].userid, buff, true)

	// receive list of peers
	var readbuff [1024]byte
	n, _ := client.Read(readbuff[:])
	msg := parseMessage(readbuff[:n])

	// save peers to own circle
	fmt.Printf("received peerlist: '%s' %v bytes\n", msg.body, n)
	if msg.body != "nil" {
		// split out peers in list
		peerlist := strings.Split(msg.body, ",")

		// for all peers in the recieved list
		for _, peer := range peerlist {
			// split address into userid, ip, and port
			peerAddr := strings.Split(peer, ":")
			peerPort, _ := strconv.Atoi(peerAddr[len(peerAddr)-1])

			// if peer doesn't exist add it
			// TODO: this line is incorrect peerExists() takes a userid as a search parameter. This if will always execute
			if c.peerExists(peerAddr[0]) == -1 {
				c.peers = append(c.peers, addPeer(peerAddr[0], peerAddr[1], peerPort))

				// let added peer know we added them
				client2, _ := net.DialUDP("udp", nil, c.peers[len(c.peers)-1].addr)
				//client2.Write([]byte(c.peers[0].userid + delimiter + strconv.FormatBool(true) + delimiter + buff))
				c.clientWrite(client2, c.peers[len(c.peers)-1].userid, buff, true)
				client2.Close()
			}
		}
	}
	go c.heartbeat()
}

// check for peer in peerlist
func (c *circle) peerExists(userid string) int {
	for i, peer := range c.peers {
		if peer.userid == userid {
			return i
		}
	}
	return -1
}

// Send a message to all peers besides self
func (c *circle) chat() {
	// create scanner
	scanner := bufio.NewScanner(os.Stdin)

	// loop on user input
	for scanner.Scan() {
		// grab user input
		buffer := scanner.Text()

		// send message to all peers
		for _, peer := range c.peers[1:] {
			client, _ := net.DialUDP("udp", nil, peer.addr)

			// if this is a quit command, else this is a just message
			if strings.Contains(string(buffer), "/quit") {
				c.clientWrite(client, peer.userid, "/quit", true)
			} else {
				c.clientWrite(client, peer.userid, buffer, false)
			}
			client.Close()
		}
		if strings.Contains(string(buffer), "/quit") {
			return
		}
	}
}

// general client send the same message structure each time
// userid, cmdflag, message
// TODO: pass in the whole peer struct instead of sending userid
// TODO: UDPDial within clientWrite and close it when done
// BUG: the receiving end doesn't get squat
func (c *circle) clientWrite(client *net.UDPConn, userid, buffer string, flag bool) {
	i := c.peerExists(userid)

	// check if the peer exists and if they are idle before sending to them
	if i != -1 && !c.peers[i].isIdle() {
		msg := packMessage(userid, flag, buffer)

		var buff bytes.Buffer

		// BUG: this does not encode the msg struct as expected... wat the hek
		binary.Write(&buff, binary.BigEndian, msg)

		// add sent message to the offline messages ledger
		c.msgs = append(c.msgs, msg)

		// send message struct
		fmt.Printf("original buffer: %v\n", buffer)
		fmt.Printf("sent msg struct: %v\n", msg)
		fmt.Printf("sent message: %s %s %s\n", msg.userid, msg.cmdflag, msg.body)
		fmt.Printf("sent buffer: %v\n", buff)
		fmt.Printf("sent buffer.Bytes(): %v\n", buff.Bytes())
		client.Write(buff.Bytes())
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

			// send message struct format
			c.clientWrite(client, peer.userid, "/heartbeat", true)
			client.Close()
		}
		time.Sleep(5 * time.Minute)
	}
}
