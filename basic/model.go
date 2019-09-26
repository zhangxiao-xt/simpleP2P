package basic

import "net"

type Peer struct {
	ID               string
	NatAddr          *net.UDPAddr
	LocalAddr        *net.UDPAddr
	Conn             *net.UDPConn
	ConnAddr         *net.UDPAddr
	QuitConnListener chan struct{} `json:"-"`
	Ok               bool
}

type PunchTry struct {
	ID         string
	Peer1Reset bool
	Peer2Reset bool
	Attempt    int
}

type Agent interface {
	Self() *Peer
	Send(*Message, *net.UDPConn, *net.UDPAddr) error
	GetPeer(string) *Peer
	SavePeer(*Peer)
	Register(string)
	GetPunchTry(string) *PunchTry
	SavePunchTry(try *PunchTry)
	ListenOnPeerConn(*Peer)
	Listen()
}

type Message struct {
	Type    string      `json:"type"`
	PeerID  string      `json:"peer_id"`
	Error   string      `json:"error,omitempty"`
	Content interface{} `json:"content,omitempty"`
}
