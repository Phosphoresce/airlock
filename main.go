package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/therecipe/qt/widgets"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Circle struct {
	peers []*peer
	msgs  []*message
}

func NewCircle() *Circle {
	circle := &Circle{
		peers: make([]*peer, 0),
		msgs:  make([]*message, 0),
		//incoming chan Message
		//outgoing chan Message
	}

	go circle.listen()
	return circle
}

type peer struct {
	addr *net.UDPAddr
	time time.Time
	// TODO: save udp client here
	client *net.UDPConn
	Userid string
}

func newPeer(ip string, port int) *peer {
	return &peer{
		addr: &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		time: time.Now(),
		Userid: func() string {
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
		Userid: id,
	}
}

type message struct {
	Userid  string
	Cmdflag string
	Body    string
}

// parse messages from recieved byte arrays
func parseMessage(buff []byte) *message {
	msg := &message{}
	gob.NewDecoder(bytes.NewReader(buff)).Decode(&msg)
	return msg
}

// create a method to create messages to send
func packMessage(id string, flag bool, buffer string) *message {
	return &message{
		Userid:  id,
		Cmdflag: strconv.FormatBool(flag),
		Body:    buffer,
	}
}

var text *widgets.QTextEdit

func main() {
	fmt.Println("Hello, Airlock!")

	// default config values
	var target string
	username := ""
	port := 9001
	gui := false

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
		case "-g", "--gui":
			gui = true
		}
	}

	// create local circle
	c := NewCircle()

	// append peer to circle
	c.peers = append(c.peers, newPeer("127.0.0.1", port))
	fmt.Printf("userid: %v\n", c.peers[0].Userid[:])

	// crudely set username
	if username != "" {
		c.peers[0].Userid = username
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

	// execute gui or terminal chat
	if gui {
		// create widget based qt gui
		widgets.NewQApplication(len(os.Args), os.Args)

		// create a window
		window := widgets.NewQMainWindow(nil, 0)
		window.SetWindowTitle("Airlock")
		window.SetMinimumSize2(200, 200)

		// create a layout
		layout := widgets.NewQVBoxLayout()

		// create a widget and set the layout
		widget := widgets.NewQWidget(nil, 0)
		widget.SetLayout(layout)

		// create a textedit and add it to the layout
		// temporarily set globally
		text = widgets.NewQTextEdit(nil)
		text.SetReadOnly(true)
		layout.AddWidget(text, 0, 0)

		// create a lineedit and add it to the layout
		input := widgets.NewQLineEdit(nil)
		input.SetPlaceholderText("1. write something")
		layout.AddWidget(input, 0, 0)

		// create a button and add it to the layout
		button := widgets.NewQPushButton2("2. click me", nil)
		button.ConnectClicked(func(checked bool) {
			text.Append("me > " + input.Text())
			c.uiChat(input.Text())
			input.Clear()

		})
		layout.AddWidget(button, 0, 0)

		// add the widget as the central widget to the window
		window.SetCentralWidget(widget)

		// show the window
		window.Show()

		// enter the main event loop
		widgets.QApplication_Exec()
	} else {
		c.chat()
	}
}

// incoming connections should create a new peer and send an updated peer list
func (c *Circle) listen() {
	// set up udp listener
	listener, err := net.ListenUDP("udp", c.peers[0].addr)
	if err != nil {
		log.Fatal("Failed to create listener: ", err)
	}
	defer listener.Close()

	// continue listening until the application is closed
	for {
		// rezero the buffer each loop
		buff := make([]byte, 1024)

		// read from the UDP listener
		rlen, remote, err := listener.ReadFromUDP(buff)

		// check for failure, if yes, continue listening, if retrieved message process it
		if err != nil {
			log.Print("Failed to read from socket: ", err)
		} else {
			// create struct with string from []byte buffer with length specified by the length of data sent to the udp listener
			msg := parseMessage(buff[:rlen])
			fmt.Printf("received msg struct: %v\n", msg)

			// decide if the message is a chat message, or command to be processed by the system
			flag, _ := strconv.ParseBool(msg.Cmdflag)
			if !flag && rlen != 0 {
				// simply print chat messages
				c.msgs = append(c.msgs, msg)
				fmt.Printf("%s > %s\n", msg.Userid[:8], msg.Body)
				fmt.Printf("%v\n", len(c.msgs))
				//inText.SetText(msg.Userid[:8] + " > " + msg.Body)
				text.Append(msg.Userid[:8] + " > " + msg.Body)

			} else if flag && rlen != 0 {
				// if command, send it to the command engine to handle
				c.cmdEngine(msg, remote, listener)
			}
		}
	}
}

func (c *Circle) cmdEngine(msg *message, remote *net.UDPAddr, listener *net.UDPConn) {
	switch {
	case strings.Contains(msg.Body, "quit"):
		// quit
		fmt.Println("peer quitting...")

		// remove peer if it exists
		i := c.peerExists(msg.Userid)
		if i != -1 {
			fmt.Println("removing peer")
			copy(c.peers[i:], c.peers[i+1:])
			c.peers[len(c.peers)-1] = nil
			c.peers = c.peers[:len(c.peers)-1]
		}

	case strings.Contains(msg.Body, "heartbeat"):
		// stop the peer from being removed from sending list
		fmt.Println("processing heartbeat...")

		// reset peer timeout if exists
		j := c.peerExists(msg.Userid)
		if j != -1 {
			c.peers[j].time = time.Now()
			fmt.Printf("time set to: %v\n", c.peers[j].time)
		}
	default:
		// process new connect
		fmt.Printf("new connect from %v\n", remote)

		// this is a new connect, check to see if we need to send a peerlist
		if len(c.peers) > 1 {
			fmt.Println("Sending peers")
			// build the peer list to send
			peerlist := make([]string, 0)

			// grab all peers besides the local peer and lastest peer
			// TODO: can I literally send the whole struct slice???
			// TODO: send peers besides first and last
			for _, peer := range c.peers[1:] {
				// TODO: need to add the userid AND the address when sending to new connect
				peerlist = append(peerlist, peer.addr.String())
			}

			// this response should be a list of ips and ports for all peers BESIDES the currently connected peer and the local peer
			c.clientWriteTo(listener, msg.Userid, strings.Join(peerlist, ","), false, remote)
			fmt.Printf("sent peerlist: %s\n", strings.Join(peerlist, ","))
		} else {
			fmt.Println("Sending no peers")
			// no peers besides the two talking to each other now, send nil peerlist
			c.clientWriteTo(listener, msg.Userid, "nil", true, remote)
		}

		// parse out the port from message... this will get removed with standard struct
		remotePort, _ := strconv.Atoi(msg.Body)

		// check if the peer exists before adding it to the circle
		index := c.peerExists(msg.Userid)
		if index == -1 {
			c.peers = append(c.peers, addPeer(msg.Userid, remote.IP.String(), remotePort))
		} else {
			fmt.Println("sending offline messages")
			for _, item := range c.msgs {
				// send them missed messages
				// BUG: this will not preserve the names of the original sender, should send whole message struct instead of body
				// TODO: this will also check if the peer exists twice? is that an issue?.. probably want to have a move this to a different command
				c.clientWriteTo(listener, msg.Userid, item.Body, false, remote)
			}
		}
	}
}

// outgoing connections should create a new peer from specified target and receive target's peer list
func (c *Circle) connect() {
	// create udp client
	client, err := net.DialUDP("udp", nil, c.peers[1].addr)
	if err != nil {
		log.Fatal("Failed to dial target: ", err)
	}
	defer client.Close()

	// port number this client will listen on.
	buff := strconv.Itoa(c.peers[0].addr.Port)

	// send message struct format, buff is port
	c.clientWrite(client, c.peers[1].Userid, buff, true)

	// receive list of peers
	var readbuff [1024]byte
	n, _ := client.Read(readbuff[:])
	msg := parseMessage(readbuff[:n])

	// save peers to own circle
	fmt.Printf("received peerlist: '%s' %v bytes\n", msg.Body, n)
	if msg.Body != "nil" {
		// split out peers in list
		peerlist := strings.Split(msg.Body, ",")

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
				c.clientWrite(client2, c.peers[len(c.peers)-1].Userid, buff, true)
				client2.Close()
			}
		}
	}
	go c.heartbeat()
}

// check for peer in peerlist
func (c *Circle) peerExists(userid string) int {
	for i, peer := range c.peers {
		if peer.Userid == userid {
			return i
		}
	}
	return -1
}

// Send a message to all peers besides self
func (c *Circle) chat() {
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
				c.clientWrite(client, peer.Userid, "/quit", true)
			} else {
				c.clientWrite(client, peer.Userid, buffer, false)
			}
			client.Close()
		}
		if strings.Contains(string(buffer), "/quit") {
			return
		}
	}
}

// Send a message to all peers besides self
func (c *Circle) uiChat(buffer string) {
	// send message to all peers
	for _, peer := range c.peers[1:] {
		client, _ := net.DialUDP("udp", nil, peer.addr)

		// if this is a quit command, else this is a just message
		if strings.Contains(buffer, "/quit") {
			c.clientWrite(client, peer.Userid, "/quit", true)
		} else {
			c.clientWrite(client, peer.Userid, buffer, false)
		}
		client.Close()
	}
	if strings.Contains(buffer, "/quit") {
		return
	}
}

// general client send the same message structure each time
// userid, cmdflag, message
// TODO: pass in the whole peer struct instead of sending userid
// TODO: UDPDial within clientWrite and close it when done
// NOTE: initiating the client and destroying it the scope of a write does not work when responses are given to an active UDPConn
// TODO: allow this clientWrite to take optionally take a remote address and call WriteTo(message, remote)
// TODO: to allow for sending of a slice of structs, need to accept interface{} type instead of string
func (c *Circle) clientWrite(client *net.UDPConn, userid, buffer string, flag bool) {
	i := c.peerExists(userid)

	// check if the peer exists and if they are idle before sending to them
	if i != -1 && !c.peers[i].isIdle() {
		msg := packMessage(userid, flag, buffer)

		var buff bytes.Buffer

		// encode struct to binary to send
		fmt.Printf("sent msg struct: %v\n", msg)
		gob.NewEncoder(&buff).Encode(&msg)

		// add sent message to the offline messages ledger
		c.msgs = append(c.msgs, msg)

		// send message struct
		client.Write(buff.Bytes())
	}
}

func (c *Circle) clientWriteTo(client *net.UDPConn, userid, buffer string, flag bool, remote *net.UDPAddr) {
	msg := packMessage(userid, flag, buffer)

	var buff bytes.Buffer

	// encode struct to binary to send
	fmt.Printf("sent msg struct: %v\n", msg)
	gob.NewEncoder(&buff).Encode(&msg)

	// add sent message to the offline messages ledger
	c.msgs = append(c.msgs, msg)

	// send message struct
	client.WriteTo(buff.Bytes(), remote)
}

func (p *peer) isIdle() bool {
	// before I send to anybody I need to make sure they are not idle
	// essentially means peer.time - time.Now() < 10 minutes
	return time.Since(p.time) > (10 * time.Minute)
}

func (c *Circle) heartbeat() {
	// want to send a control message at a designated interval
	// timer 5 minutes
	for {
		for _, peer := range c.peers[1:] {
			client, _ := net.DialUDP("udp", nil, peer.addr)

			// send message struct format
			c.clientWrite(client, peer.Userid, "/heartbeat", true)
			client.Close()
		}
		time.Sleep(5 * time.Minute)
	}
}
