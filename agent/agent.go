package agent

import (
	"encoding/json"
	"fmt"
	"github.com/HenryTank/simpleP2P/basic"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

//Find ip of privileged local network interface
func LocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func NewUDPConn(port int) (*net.UDPConn, *net.UDPAddr) {

	if port < 10000 {
		port = rand.New(rand.NewSource(time.Now().UnixNano())).Intn(65535-10000) + 10000
	}

	localAddr, _ := net.ResolveUDPAddr("udp", LocalIP()+":"+strconv.Itoa(port))
	conn, e := net.ListenUDP("udp", localAddr)
	if e != nil {
		return nil, nil
	}
	return conn, localAddr
}

//Dispatch messages to handle functions according to message types
func HandleMessage(buf []byte, peerAddr *net.UDPAddr, agent basic.Agent) {

	var msg basic.Message

	e := json.Unmarshal(buf, &msg)
	if e != nil {
		log.Println(e)
		return
	}

	fmt.Println("--MSG---------------------")
	fmt.Printf("Type: %s, PeerID: %s, PeerAddr: %s, Content: %v\n\n", msg.Type, msg.PeerID, peerAddr.String(), msg.Content)

	switch msg.Type {
	case "REG":
		handleRegisterMsg(&msg, peerAddr, agent)
	case "REG-RES":
		handleRegisterResponseMsg(&msg, peerAddr, agent)
	case "PUNCH-REQ":
		handleRequestEstablishMsg(&msg, peerAddr, agent)
	case "PUNCH-COMMAND":
		handlePunchCommandMsg(&msg, peerAddr, agent)
	case "PUNCH-FAIL":
		handlePunchFailMsg(&msg, peerAddr, agent)
	case "PUNCH":
		handlePunchMsg(&msg, peerAddr, agent)
	case "PUNCH-RES":
		handlePunchResponseMsg(&msg, peerAddr, agent)
	case "NORMAL":
		handleNormalMsg(&msg, peerAddr, agent)
	}

}

//Handle register request from client
func handleRegisterMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	//Decode msg content
	c := struct {
		LocalAddr *net.UDPAddr
	}{}
	b, _ := json.Marshal(msg.Content)
	_ = json.Unmarshal(b, &c)

	peer := &basic.Peer{
		ID:        msg.PeerID,
		NatAddr:   addr,
		LocalAddr: c.LocalAddr,
		Conn:      agent.Self().Conn,
		ConnAddr:  addr,
	}

	agent.SavePeer(peer)

	_ = agent.Send(&basic.Message{
		Type:    "REG-RES",
		PeerID:  agent.Self().ID,
		Content: peer,
	}, agent.Self().Conn, peer.NatAddr)

}

//Handle register response from server
func handleRegisterResponseMsg(msg *basic.Message, _ *net.UDPAddr, agent basic.Agent) {

	b, _ := json.Marshal(msg.Content)
	var c basic.Peer
	_ = json.Unmarshal(b, &c)

	agent.Self().NatAddr = c.NatAddr

}

//Do send punch commands to clients
func conductPunchProcess(peer1ID string, peer2ID string, agent basic.Agent) {

	//Decide punch sequence
	var punchID string
	var peer1, peer2 *basic.Peer
	if peer1ID < peer2ID {
		punchID = peer1ID + "-" + peer2ID
		peer1 = agent.GetPeer(peer1ID)
		peer2 = agent.GetPeer(peer2ID)
	} else {
		punchID = peer2ID + "-" + peer1ID
		peer1 = agent.GetPeer(peer2ID)
		peer2 = agent.GetPeer(peer1ID)
	}

	if peer1 == nil || peer2 == nil {
		fmt.Println("Error: peer not register!")
		return
	}

	var try = agent.GetPunchTry(punchID)
	if !try.Peer1Reset || !try.Peer2Reset {
		return
	}
	defer func() {
		try.Attempt++
		try.Peer1Reset, try.Peer2Reset = false, false
		agent.SavePunchTry(try)
	}()

	var tmpPeer *basic.Peer
	if (try.Attempt % 2) == 1 {
		tmpPeer = peer1
		peer1 = peer2
		peer2 = tmpPeer
	}

	//Send punch commands
	_ = agent.Send(&basic.Message{
		Type:    "PUNCH-COMMAND",
		PeerID:  agent.Self().ID,
		Content: peer2,
	}, agent.Self().Conn, peer1.NatAddr)

	time.Sleep(time.Second * 3)

	_ = agent.Send(&basic.Message{
		Type:    "PUNCH-COMMAND",
		PeerID:  agent.Self().ID,
		Content: peer1,
	}, agent.Self().Conn, peer2.NatAddr)

}

//Handle punch request from client for connecting another client
func handleRequestEstablishMsg(msg *basic.Message, _ *net.UDPAddr, agent basic.Agent) {

	//Decode msg content
	c := struct {
		PeerID string
	}{}
	b, _ := json.Marshal(msg.Content)
	_ = json.Unmarshal(b, &c)

	//Conduct punch try
	conductPunchProcess(msg.PeerID, c.PeerID, agent)

}

//Handle punch fail response from client
func handlePunchFailMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	//Decode msg content
	c := struct {
		PeerID       string
		LocalAddress *net.UDPAddr
	}{}
	b, _ := json.Marshal(msg.Content)
	_ = json.Unmarshal(b, &c)

	//Reset peer
	peer := &basic.Peer{
		ID:        msg.PeerID,
		NatAddr:   addr,
		LocalAddr: c.LocalAddress,
	}
	agent.SavePeer(peer)

	//Update punch try info
	var punchID string
	var peer1Reset, peer2Reset = false, false
	if msg.PeerID < c.PeerID {
		punchID = msg.PeerID + "-" + c.PeerID
		peer1Reset = true
	} else {
		punchID = c.PeerID + "-" + msg.PeerID
		peer2Reset = true
	}

	try := agent.GetPunchTry(punchID)
	try.Peer1Reset = try.Peer1Reset || peer1Reset
	try.Peer2Reset = try.Peer2Reset || peer2Reset
	agent.SavePunchTry(try)

	//Conduct next punch try if necessary
	if try.Peer1Reset && try.Peer2Reset {
		conductPunchProcess(msg.PeerID, c.PeerID, agent)
	}

}

//Handle punch command from server
func handlePunchCommandMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	//Decode msg content
	c := struct {
		ID        string
		NatAddr   *net.UDPAddr
		LocalAddr *net.UDPAddr
	}{}
	b, _ := json.Marshal(msg.Content)
	_ = json.Unmarshal(b, &c)

	//Peer prepare
	peer := agent.GetPeer(c.ID)
	if peer == nil {
		peer = &basic.Peer{
			ID:   c.ID,
			Conn: agent.Self().Conn,
			Ok:   false,
		}
	}
	peer.NatAddr = c.NatAddr
	peer.LocalAddr = c.LocalAddr
	agent.SavePeer(peer)

	//Do punch connection
	for i := 0; i < 10; i++ {

		_ = agent.Send(&basic.Message{
			Type:   "PUNCH",
			PeerID: agent.Self().ID,
		}, peer.Conn, peer.NatAddr)

		time.Sleep(time.Second * 1)

		if p := agent.GetPeer(peer.ID); p.Ok {

			_ = agent.Send(&basic.Message{
				Type:    "NORMAL",
				PeerID:  agent.Self().ID,
				Content: "Hi, my friend!",
			}, p.Conn, p.ConnAddr)

			return
		}

	}

	//Reset peer conn
	peer = agent.GetPeer(c.ID)

	if peer.Conn != agent.Self().Conn {
		_ = peer.Conn.Close()
		close(peer.QuitConnListener)
	}

	newConn, localAddr := NewUDPConn(-1)
	peer.Conn = newConn
	go agent.ListenOnPeerConn(peer)

	agent.SavePeer(peer)

	//Send fail response
	_ = agent.Send(&basic.Message{
		Type:   "PUNCH-FAIL",
		PeerID: agent.Self().ID,
		Content: struct {
			PeerID       string
			LocalAddress *net.UDPAddr
		}{
			PeerID:       peer.ID,
			LocalAddress: localAddr,
		},
	}, peer.Conn, addr)

}

//Handle punch message from another client
func handlePunchMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	peer := agent.GetPeer(msg.PeerID)
	if peer == nil {
		return
	}
	peer.ConnAddr = addr
	agent.SavePeer(peer)

	_ = agent.Send(&basic.Message{
		Type:   "PUNCH-RES",
		PeerID: agent.Self().ID,
	}, peer.Conn, peer.ConnAddr)
}

//Handle punch response message from another client
func handlePunchResponseMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	peer := agent.GetPeer(msg.PeerID)
	peer.Ok = true
	peer.ConnAddr = addr
	agent.SavePeer(peer)

}

func handleNormalMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

}
