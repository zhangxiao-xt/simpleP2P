package agent

import (
	"encoding/json"
	"github.com/HenryTank/simpleP2P/basic"
	"net"
	"time"
)

//Handle register response from server
func handleRegisterResponseMsg(msg *basic.Message, _ *net.UDPAddr, agent basic.Agent) {

	b, _ := json.Marshal(msg.Content)
	var c basic.Peer
	_ = json.Unmarshal(b, &c)

	agent.Self().NatAddr = c.NatAddr

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
