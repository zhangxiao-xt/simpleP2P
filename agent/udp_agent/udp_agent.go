package udp_agent

import (
	"encoding/json"
	"fmt"
	"github.com/HenryTank/simpleP2P/agent"
	"github.com/HenryTank/simpleP2P/basic"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type Agent struct {
	selfPeer    *basic.Peer
	remotePeers map[string]*basic.Peer
	serverPeer  *basic.Peer
	conn        *net.UDPConn
}

func New(localPort int) *Agent {

	//self peer
	if localPort < 10000 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		localPort = r.Intn(65535-10000) + 10000
	}
	address, e := net.ResolveUDPAddr("udp", localIP()+":"+strconv.Itoa(localPort))
	if e != nil {
		log.Fatal(e)
	}

	selfPeer := &basic.Peer{
		ID:        primitive.NewObjectID().Hex(),
		LocalAddr: address,
	}

	//remote peers
	remotePeers := make(map[string]*basic.Peer)

	//conn
	conn, e := net.ListenUDP("udp", selfPeer.LocalAddr)
	if e != nil {
		log.Fatal(e)
	}

	return &Agent{
		selfPeer:    selfPeer,
		remotePeers: remotePeers,
		conn:        conn,
	}
}

func (c *Agent) Listen() {
	fmt.Printf("ID: %s \nlistening udp on %s\n", c.Self().ID, c.Self().LocalAddr.String())
	for {

		buf := make([]byte, 2048)

		e := c.conn.SetDeadline(time.Now().Add(time.Second * 5))
		if e != nil {

		}

		n, address, e := c.conn.ReadFromUDP(buf)
		if e != nil {
			continue
		}

		agent.HandleMessage(buf[:n], address, c)
	}
}

func (c *Agent) Send(msg *basic.Message, addr *net.UDPAddr) error {
	buf, _ := json.Marshal(msg)
	_, e := c.conn.WriteToUDP(buf, addr)
	fmt.Println(addr)
	if e != nil {
		fmt.Println(e)
		return e
	}
	return nil
}

func (c *Agent) SetSelfNatAddr(addr *net.UDPAddr)  {
	c.selfPeer.LocalAddr = addr
}

func localIP() string {
	conn, err := net.Dial("udp", "112.126.117.107:23311")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func (c *Agent) RegisterToServer(serverUrl string) {

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
		Content: basic.RegContent{
			Address: c.Self().LocalAddr,
		},
	}, c.serverPeer.NatAddr)

	fmt.Println(e)

}

func (c *Agent) ConnectToPeer(peerID string) {
	_ = c.Send(&basic.Message{
		Type:   "REQ-EST",
		PeerID: c.Self().ID,
		Content: basic.ReqEstContent{
			PeerID: peerID,
		},
	}, c.serverPeer.NatAddr)
}

func (c *Agent) SavePeer(peer *basic.Peer) {
	fmt.Println(peer)
	c.remotePeers[peer.ID] = peer
}

func (c *Agent) GetPeer(peerID string) *basic.Peer {
	peer, ok := c.remotePeers[peerID]
	if !ok {
		return nil
	}
	return peer
}

func (c *Agent) Self() *basic.Peer {
	return c.selfPeer
}
