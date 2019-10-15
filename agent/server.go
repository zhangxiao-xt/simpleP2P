package agent

import (
	"encoding/json"
	"fmt"
	"github.com/HenryTank/simpleP2P/basic"
	"net"
	"time"
)

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

	if try.Attempt == 0 {
		_ = agent.Send(&basic.Message{
			Type:    "PUNCH-COMMAND",
			PeerID:  agent.Self().ID,
			Content: peer2,
		}, agent.Self().Conn, peer1.NatAddr)

		_ = agent.Send(&basic.Message{
			Type:    "PUNCH-COMMAND",
			PeerID:  agent.Self().ID,
			Content: peer1,
		}, agent.Self().Conn, peer2.NatAddr)

		return
	}

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
