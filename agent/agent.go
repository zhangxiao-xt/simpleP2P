package agent

import (
	"encoding/json"
	"fmt"
	"github.com/HenryTank/simpleP2P/basic"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net"
	"time"
)

type Agent struct {
	selfPeer    *basic.Peer
	remotePeers map[string]*basic.Peer
	serverPeer  *basic.Peer
	conn        *net.UDPConn
	punchTries  map[string]*basic.PunchTry
}

func New(localPort int) *Agent {

	//self peer
	conn, address := NewUDPConn(localPort)

	selfPeer := &basic.Peer{
		ID:        primitive.NewObjectID().Hex(),
		LocalAddr: address,
		Conn:      conn,
	}

	return &Agent{
		selfPeer:    selfPeer,
		remotePeers: make(map[string]*basic.Peer),
		punchTries:  make(map[string]*basic.PunchTry),
		conn:        conn,
	}
}

func (c *Agent) Register(serverUrl string) {

	serverAddr, e := net.ResolveUDPAddr("udp", serverUrl)
	if e != nil {
		log.Fatal(e)
	}

	c.serverPeer = &basic.Peer{
		ID:      "server",
		NatAddr: serverAddr,
	}

	e = c.Send(&basic.Message{
		Type:   "REG",
		PeerID: c.Self().ID,
		Content: struct {
			LocalAddr *net.UDPAddr
		}{
			LocalAddr: c.Self().LocalAddr,
		},
	}, c.Self().Conn, c.serverPeer.NatAddr)

}

func (c *Agent) Self() *basic.Peer {
	return c.selfPeer
}

func (c *Agent) Send(msg *basic.Message, conn *net.UDPConn, addr *net.UDPAddr) error {
	buf, _ := json.Marshal(msg)
	_, e := conn.WriteToUDP(buf, addr)
	if e != nil {
		return e
	}
	return nil
}

func (c *Agent) Listen() {
	fmt.Printf("ID: %s \nlistening udp on %s\n", c.Self().ID, c.Self().LocalAddr.String())
	for {
		c.readAndProcessMessage(c.conn)
	}
}

func (c *Agent) ListenOnPeerConn(peer *basic.Peer) {

	fmt.Printf("!!! peer listen start\n")

	if peer.Conn == nil {
		return
	}

	peer.Meta.ListenerQuit = make(chan struct{})
	for {
		select {
		case <-peer.Meta.ListenerQuit:
			fmt.Println("!!! close listener")
			break
		default:
			c.readAndProcessMessage(peer.Conn)
		}
	}

}

func (c *Agent) readAndProcessMessage(conn *net.UDPConn) {

	buf := make([]byte, 2048)

	e := conn.SetDeadline(time.Now().Add(time.Second * 5))
	if e != nil {

	}

	n, address, e := conn.ReadFromUDP(buf)
	if e != nil {
		return
	}

	go handleMessage(buf[:n], address, c)
}

func (c *Agent) GetPeer(peerID string) *basic.Peer {
	peer, ok := c.remotePeers[peerID]
	if !ok {
		return nil
	}
	return peer
}

func (c *Agent) SavePeer(peer *basic.Peer) {
	c.remotePeers[peer.ID] = peer
}

func (c *Agent) ConnectToPeer(peerID string) {
	_ = c.Send(&basic.Message{
		Type:   "PUNCH-REQ",
		PeerID: c.Self().ID,
		Content: struct {
			PeerID string
		}{
			PeerID: peerID,
		},
	}, c.Self().Conn, c.serverPeer.NatAddr)
}

func (c *Agent) GetPunchTry(punchID string) *basic.PunchTry {
	try, ok := c.punchTries[punchID]
	if !ok {
		c.punchTries[punchID] = &basic.PunchTry{
			ID:         punchID,
			Peer1Reset: true,
			Peer2Reset: true,
			Attempt:    0,
		}
		return c.punchTries[punchID]
	}
	return try
}

func (c *Agent) SavePunchTry(try *basic.PunchTry) {
	c.punchTries[try.ID] = try
}

//Dispatch messages to handle functions according to message types
func handleMessage(buf []byte, peerAddr *net.UDPAddr, agent basic.Agent) {

	var msg basic.Message

	e := json.Unmarshal(buf, &msg)
	if e != nil {
		log.Println(e)
		return
	}

	fmt.Println("--MSG---------------------")
	fmt.Printf("Type: %s, PeerID: %s, PeerAddr: %s, Content: %v\n\n", msg.Type, msg.PeerID, peerAddr.String(), msg.Content)

	switch msg.Type {

	//server
	case "REG":
		handleRegisterMsg(&msg, peerAddr, agent)
	case "PUNCH-REQ":
		handleRequestEstablishMsg(&msg, peerAddr, agent)
	case "PUNCH-FAIL":
		handlePunchFailMsg(&msg, peerAddr, agent)

	//client
	case "REG-RES":
		handleRegisterResponseMsg(&msg, peerAddr, agent)
	case "PUNCH-COMMAND":
		handlePunchCommandMsg(&msg, peerAddr, agent)
	case "PUNCH":
		handlePunchMsg(&msg, peerAddr, agent)
	case "PUNCH-RES":
		handlePunchResponseMsg(&msg, peerAddr, agent)
	case "NORMAL":
		handleNormalMsg(&msg, peerAddr, agent)

	}

}
