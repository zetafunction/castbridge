package main

import (
	"flag"
	"github.com/zetafunction/castbridge/discovery"
	"github.com/zetafunction/castbridge/forwarder"
	"log"
)

// FIXME: Allow port to be specified.
var endpointFlag = flag.String("endpoint", "", "")

func main() {
	flag.Parse()

	if *endpointFlag == "" {
		log.Print("starting CastBridge in service mode")
		forwarder.Listen()
	} else {
		log.Print("starting CastBridge in client mode")
		discovery.ListenForDIAL(*endpointFlag)
	}
}
