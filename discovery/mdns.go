package discovery

import (
	"bytes"
	"encoding/binary"
	"github.com/zetafunction/castbridge/forwarder"
	"log"
	"net"
)

type ForwardingMode int

const (
	MDNSQueryForwarding  = iota
	MDNSAnswerForwarding = iota
)

var mDNSMulticastAddr = &net.UDPAddr{
	IP:   net.ParseIP("224.0.0.251"),
	Port: 5353,
}

type dnsHeader struct {
	Id                    uint16
	Flags                 uint16
	QuestionCount         uint16
	AnswerCount           uint16
	RecordCount           uint16
	AdditionalRecordCount uint16
}

// A cast query will only have one question record and one answer record.
var expectedCastQueryHeader dnsHeader = dnsHeader{
	0x0,
	0x0,
	1,
	1,
	0,
	0,
}

// A cast answer will have one answer record and three additional records.
var expectedCastAnswerHeader dnsHeader = dnsHeader{
	0x0,
	0x8400,
	0,
	1,
	0,
	3,
}

// The FQDN in the data section is encoded as a series of Pascal-style strings. Each string starts
// with a byte length followed by that number of UTF-8 characters. A string with a byte length of
// zero terminates the FQDN.
func parseFQDNFromData(reader *bytes.Reader) (string, error) {
	var fqdn string
	var length byte
	for {
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			return fqdn, err
		}
		if length == 0 {
			return fqdn, nil
		}
		component := make([]byte, length)
		if err := binary.Read(reader, binary.BigEndian, &component); err != nil {
			return fqdn, err
		}
		fqdn += string(component)
		fqdn += "."
	}
}

func ListenForMDNS(mode ForwardingMode, clientChannel chan<- *forwarder.Request) {
	if mode != MDNSQueryForwarding && mode != MDNSAnswerForwarding {
		panic("unknown ForwardingMode specified")
	}

	conn, err := net.ListenMulticastUDP("udp", nil, mDNSMulticastAddr)
	if err != nil {
		log.Fatalf("net.ListenMulticastUDP failed: %v", err)
	}

	for {
		buf := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("UDPConn.ReadFromUDP failed: %v", err)
		}

		reader := bytes.NewReader(buf[:n])

		var header dnsHeader
		if err := binary.Read(reader, binary.BigEndian, &header); err != nil {
			log.Printf("failed to unpack DNS header: %v", err)
			continue
		}

		if (mode == MDNSQueryForwarding && header != expectedCastQueryHeader) ||
			(mode == MDNSAnswerForwarding && header != expectedCastAnswerHeader) {
			continue
		}

		fqdn, err := parseFQDNFromData(reader)
		if err != nil {
			log.Printf("parseFQDNFromAnswer failed: %v", err)
			continue
		}
		if fqdn != "_googlecast._tcp.local." {
			continue
		}

		if header == expectedCastQueryHeader {
			log.Printf("got cast query header")
		} else {
			log.Printf("got cast answer header")
		}

		// TODO: Implement forwarding strategy. Unfortunately, the simple forwarder isn't
		// sufficient, since mDNS responses are multicast. This introduces two problems:
		// - The forwarder currently assumes a request-response model, but if the reply is
		//   multicasted, it has no easy way of associating a response with its request.
		// - Simply rebroadcasting it on the subnet where the query originated is likely to
		//   lead to a multicast loop on many systems.
	}
}
