package agent

import (
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
