package main

import (
	"bufio"
	"flag"
	"github.com/HenryTank/simpleP2P/agent/udp_agent"
	"os"
	"strings"
)

func main() {

	port := flag.Int("port", -1, "local port")
	server := flag.String("server", "", "server address")
	flag.Parse()
	//server := flag.String("server", "47.107.166.228:25728", "server address")
	//112.126.117.107:23311

	agent := udp_agent.New(*port)

	if *server != "" {
		go agent.Listen()
		agent.Register(*server)
		//agent.Register("112.126.117.107:23311")
		//agent.Register("39.97.128.62:23311")
		//agent.Register("112.126.119.244:23311")
		for {
			buf := bufio.NewReader(os.Stdin)
			sentence, _ := buf.ReadBytes('\n')
			peerID := strings.TrimSpace(string(sentence))
			//fmt.Println(peerID)
			agent.ConnectToPeer(peerID)

			////for {
			//
			//	addr, _ := net.ResolveUDPAddr("udp", peerID)
			//
			//	_ = agent.Send(&basic.Message{
			//		Type:   "CONN",
			//		PeerID: agent.Self().ID,
			//	}, addr)
			//
			//	//time.Sleep(time.Second * 5)
			////}
		}
	} else {
		agent.Listen()
	}

}
