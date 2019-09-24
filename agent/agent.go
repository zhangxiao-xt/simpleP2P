package agent

import (
	"encoding/json"
	"fmt"
	"github.com/HenryTank/simpleP2P/basic"
	"net"
	"time"
)

func HandleMessage(buf []byte, peerAddr *net.UDPAddr, agent basic.Agent) {

	var msg basic.Message

	e := json.Unmarshal(buf, &msg)
	if e != nil {
		//log.Println(e)
		return
	}

	fmt.Println("--MSG---------------------")
	fmt.Printf("Type: %s, PeerID: %s, PeerAddr: %s, Content: ", msg.Type, msg.PeerID, peerAddr.String())
	fmt.Println(msg.Content)

	switch msg.Type {
	case "REG":
		handleRegisterMsg(&msg, peerAddr, agent)
	case "REG-RES":
		handleRegisterResponseMsg(&msg, peerAddr, agent)
	case "REQ-EST":
		handleRequestEstablishMsg(&msg, peerAddr, agent)
	case "EST":
		handleEstablishMsg(&msg, peerAddr, agent)
	case "CONN":
		handleConnectionMsg(&msg, peerAddr, agent)
	case "NORMAL":
		handleNormalMsg(&msg, peerAddr, agent)
	}

}

func handleRegisterMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

	fmt.Println(msg)

	b, _ := json.Marshal(msg.Content)
	var c basic.RegContent
	_ = json.Unmarshal(b, &c)

	peer := &basic.Peer{
		ID:        msg.PeerID,
		NatAddr:   peerAddr,
		LocalAddr: c.Address,
	}

	agent.SavePeer(peer)

	_ = agent.Send(&basic.Message{
		Type:    "REG-RES",
		PeerID:  agent.Self().ID,
		Content: peer,
	}, peer.NatAddr)

}

func handleRegisterResponseMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent)  {

	b, _ := json.Marshal(msg.Content)
	var c basic.Peer
	_ = json.Unmarshal(b, &c)

	agent.Self().NatAddr = c.NatAddr

	fmt.Println(c.NatAddr)

}

func handleRequestEstablishMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

	peer := agent.GetPeer(msg.PeerID)

	if peer == nil || peer.NatAddr.String() != peerAddr.String() {
		fmt.Println("peer not register!!")
		return
	}

	b, _ := json.Marshal(msg.Content)
	var c basic.ReqEstContent
	_ = json.Unmarshal(b, &c)

	targetPeer := agent.GetPeer(c.PeerID)

	fmt.Println("--H:ESTREQ---------------------")
	fmt.Println(c.PeerID)
	fmt.Println("peer: ", peer)
	fmt.Println("targetPeer: ", targetPeer)

	_ = agent.Send(&basic.Message{
		Type:    "EST",
		PeerID:  agent.Self().ID,
		Content: targetPeer,
	}, peer.NatAddr)

	_ = agent.Send(&basic.Message{
		Type:    "EST",
		PeerID:  agent.Self().ID,
		Content: peer,
	}, targetPeer.NatAddr)

}

func handleEstablishMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

	b, _ := json.Marshal(msg.Content)
	var c basic.EstContent
	_ = json.Unmarshal(b, &c)

	peer := &basic.Peer{
		ID:        c.ID,
		NatAddr:   c.NatAddr,
		LocalAddr: c.LocalAddr,
	}

	fmt.Println("he11111")

	go func() {
		for {

			fmt.Println("he222222")

			var peerAddr = peer.NatAddr;

			if agent.Self().NatAddr.IP.String() == peer.NatAddr.IP.String() {
				peerAddr = peer.LocalAddr
			}

			fmt.Println("he333333")

			_ = agent.Send(&basic.Message{
				Type:   "CONN",
				PeerID: agent.Self().ID,
			}, peerAddr)


			//agent.RegisterToServer("112.126.117.107:23311")

			time.Sleep(time.Second * 3)

			if p := agent.GetPeer(peer.ID); p != nil {

				_ = agent.Send(&basic.Message{
					Type:    "NORMAL",
					PeerID:  agent.Self().ID,
					Content: "Hi my friend",
				}, peerAddr)

				fmt.Println("hi~~~~~~")

				break
			}

		}
	}()

}

func handleConnectionMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

	peer := &basic.Peer{
		ID:      msg.PeerID,
		NatAddr: peerAddr,
	}

	agent.SavePeer(peer)

}

func handleNormalMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

}
