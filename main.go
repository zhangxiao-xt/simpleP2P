package main

import (
	"bufio"
	"flag"
	"github.com/HenryTank/simpleP2P/agent/udp_agent"
	"github.com/HenryTank/simpleP2P/basic"
	"net"
	"os"
	"strings"
)

func main() {

	port := flag.Int("port", 46129, "local port")
	server := flag.String("server", "112.126.117.107:23311", "server address")
	//server := flag.String("server", "47.107.166.228:25728", "server address")
	flag.Parse()

	agent := udp_agent.New(*port)

	//if *port > 1000 {
	//	agent.Listen()
	//}

	go agent.Listen()

	if *server != "" {
		agent.RegisterToServer(*server)
	}

	//agent.RegisterToServer("112.126.117.107:23311")
	//agent.RegisterToServer("39.97.128.62:23311")
	//agent.RegisterToServer("112.126.119.244:23311")

	for {
		buf := bufio.NewReader(os.Stdin)
		sentence, _ := buf.ReadBytes('\n')
		peerID := strings.TrimSpace(string(sentence))
		//fmt.Println(peerID)
		//agent.ConnectToPeer(peerID)

		//for {

			addr, _ := net.ResolveUDPAddr("udp", peerID)

			_ = agent.Send(&basic.Message{
				Type:   "CONN",
				PeerID: agent.Self().ID,
			}, addr)

			//time.Sleep(time.Second * 5)
		//}
	}


}