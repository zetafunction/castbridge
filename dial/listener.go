package dial

import (
	"bufio"
	"bytes"
	"github.com/zetafunction/castbridge/forwarder"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

const DIALServiceSearchType = "urn:dial-multiscreen-org:service:dial:1"

var SSDPAddress = &net.UDPAddr{
	IP:   net.ParseIP("239.255.255.250"),
	Port: 1900,
}

func Listen(method string, rpcServerAddr string) {
	client, err := rpc.Dial("tcp", rpcServerAddr+":8714")
	if err != nil {
		log.Fatalf("rpc.Dial failed: %v", err)
	}

	conn, err := net.ListenMulticastUDP("udp", nil, SSDPAddress)
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

		if req.Header.Get(http.CanonicalHeaderKey("ST")) != DIALServiceSearchType {
			continue
		}

		if req.Method != method {
			continue
		}

		log.Printf("Got a SSDP discovery request, forwarding...")
		respBuf := make([]byte, 1024)
		call := client.Go("Forwarder.Forward", &forwarder.Args{}, &respBuf, nil)
		go handleRPCResponse(call, addr)
	}
}

func handleRPCResponse(call *rpc.Call, addr *net.UDPAddr) {
	<-call.Done
	if call.Error != nil {
		log.Printf("RPC failed: %v", call.Error)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("net.DialUDP failed: %v", err)
		return
	}
	defer conn.Close()

	reply := call.Reply.([]byte)
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
