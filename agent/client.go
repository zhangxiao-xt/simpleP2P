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
	isFirstPunch := false
	if peer == nil {
		isFirstPunch = true
		peer = &basic.Peer{
			ID:   c.ID,
			Conn: agent.Self().Conn,
			Meta: &basic.PeerMeta{
				IngressOk: false,
				EgressOk:  false,
			},
		}
	}
	peer.NatAddr = c.NatAddr
	peer.LocalAddr = c.LocalAddr
	agent.SavePeer(peer)

	//Try local address if it is the first time punch
	if isFirstPunch {
		for i := 10; i > 0; i-- {

			_ = agent.Send(&basic.Message{
				Type:   "PUNCH",
				PeerID: agent.Self().ID,
			}, peer.Conn, peer.LocalAddr)

			time.Sleep(time.Millisecond * 100)

			p := agent.GetPeer(peer.ID)

			if p.Meta.IngressOk {
				i = 30
			}

			if p.Meta.EgressOk {
				return
			}

		}
	} else {

		//Do punch connection
		for i := 0; i < 10; i++ {

			_ = agent.Send(&basic.Message{
				Type:   "PUNCH",
				PeerID: agent.Self().ID,
			}, peer.Conn, peer.NatAddr)

			time.Sleep(time.Millisecond * 400)

			p := agent.GetPeer(peer.ID)

			if p.Meta.IngressOk {
				i = 30
			}

			if p.Meta.EgressOk {
				return
			}

		}

	}



	//Reset peer conn
	peer = agent.GetPeer(c.ID)

	if peer.Conn != agent.Self().Conn {
		_ = peer.Conn.Close()
		close(peer.Meta.ListenerQuit)
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
	peer.Meta.IngressOk = true
	agent.SavePeer(peer)

	_ = agent.Send(&basic.Message{
		Type:   "PUNCH-RES",
		PeerID: agent.Self().ID,
	}, peer.Conn, peer.ConnAddr)
}

//Handle punch response message from another client
func handlePunchResponseMsg(msg *basic.Message, addr *net.UDPAddr, agent basic.Agent) {

	peer := agent.GetPeer(msg.PeerID)
	peer.Meta.EgressOk = true
	peer.ConnAddr = addr
	agent.SavePeer(peer)

}

func handleNormalMsg(msg *basic.Message, peerAddr *net.UDPAddr, agent basic.Agent) {

}
