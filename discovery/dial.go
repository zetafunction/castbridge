package discovery

import (
	"bufio"
	"bytes"
	"github.com/zetafunction/castbridge/forwarder"
	"log"
	"net"
	"net/http"
	"time"
)

const dialServiceSearchType = "urn:dial-multiscreen-org:service:dial:1"

var ssdpMulticastAddr = &net.UDPAddr{
	IP:   net.ParseIP("239.255.255.250"),
	Port: 1900,
}

func ListenForDIAL(clientChannel chan<- *forwarder.Request) {
	conn, err := net.ListenMulticastUDP("udp", nil, ssdpMulticastAddr)
	if err != nil {
		log.Fatalf("net.ListenMulticastUDP failed: %v", err)
	}

	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("UDPConn.ReadFromUDP failed: %v", err)
		}

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf[:n])))
		if err != nil {
			log.Printf("got malformed request from %v: %v", addr, err)
			continue
		}

		if req.Header.Get(http.CanonicalHeaderKey("ST")) != dialServiceSearchType {
			continue
		}

		if req.Method != "M-SEARCH" {
			continue
		}

		clientChannel <- &forwarder.Request{
			&forwarder.Args{
				ssdpMulticastAddr,
				buf[:n],
			},
			newReplyHandler(addr),
		}
	}
}

func newReplyHandler(addr *net.UDPAddr) func([]byte) {
	return func(reply []byte) {
		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			log.Printf("net.DialUDP failed: %v", err)
			return
		}
		defer conn.Close()

		// TODO: Arbitrary deadline. Should probably be configurable.
		if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			log.Printf("UDPConn.SetDeadline failed: %v", err)
			return
		}
		n, err := conn.Write(reply)
		if err != nil {
			log.Printf("UDPConn.Write failed: %v", err)
			return
		}
		if n != len(reply) {
			log.Printf("UDPConn.Write failed: short write: %v", n)
			return

		}
	}
}
