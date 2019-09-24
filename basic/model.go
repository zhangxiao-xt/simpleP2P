package basic

import "net"

type Agent interface {
	Send(*Message, *net.UDPAddr) error
	SavePeer(*Peer)
	GetPeer(string) *Peer
	Self() *Peer
	RegisterToServer(serverUrl string)
}

type Peer struct {
	ID        string
	NatAddr   *net.UDPAddr
	LocalAddr *net.UDPAddr
}

type Message struct {
	Type    string      `json:"type"`
	PeerID  string      `json:"peer_id"`
	Error   string      `json:"error,omitempty"`
	Content interface{} `json:"content,omitempty"`
}

type ReqEstContent struct {
	PeerID string `json:"peer_id"`
}

type EstContent struct {
	ID        string
	NatAddr   *net.UDPAddr
	LocalAddr *net.UDPAddr
}

type RegContent struct {
	Address *net.UDPAddr
}
