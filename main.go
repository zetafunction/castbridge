package main

import (
	"flag"
	"github.com/zetafunction/castbridge/discovery"
	"github.com/zetafunction/castbridge/forwarder"
	"log"
)

var endpointFlag = flag.String("endpoint", "", "")
var portFlag = flag.Int("port", 0, "")
var serviceFlag = flag.Bool("client", false, "")

func main() {
	flag.Parse()

	if *serviceFlag {
		log.Print("starting CastBridge in service mode")
		clientChannel := forwarder.NewClient(*endpointFlag)
		go discovery.ListenForMDNS(discovery.MDNSAnswerForwarding, clientChannel)
		go forwarder.Listen(*portFlag)
	} else {
		log.Print("starting CastBridge in client mode")
		clientChannel := forwarder.NewClient(*endpointFlag)
		go discovery.ListenForDIAL(clientChannel)
		go discovery.ListenForMDNS(discovery.MDNSQueryForwarding, clientChannel)
		go forwarder.Listen(*portFlag)
	}
	select {}
}
