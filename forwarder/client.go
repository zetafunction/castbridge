package forwarder

import (
	"log"
	"net"
	"net/rpc"
)

type Client struct {
	queue chan<- func(*rpc.Client)
}

func (c *Client) Send(addr *net.UDPAddr, buf []byte) {
	c.queue <- func(client *rpc.Client) {
		if err := client.Call("Forwarder.Send", &Args{addr, buf}, nil); err != nil {
			log.Printf("Client.Call failed: %v", err)
			return
		}
	}
}

func (c *Client) SendAndReply(addr *net.UDPAddr, buf []byte, onReply func([]byte)) {
	c.queue <- func(client *rpc.Client) {
		var reply []byte
		if err := client.Call("Forwarder.SendAndReply", &Args{addr, buf}, &reply); err != nil {
			log.Printf("Client.Call failed: %v", err)
			return
		}
		onReply(reply)
	}
}

func NewClient(endpointAddr string) *Client {
	channel := make(chan func(*rpc.Client), 100)
	client := &Client{
		channel,
	}
	go handleRequests(endpointAddr, channel)
	return client
}

func handleRequests(endpointAddr string, channel <-chan func(*rpc.Client)) {
	for rpcCall := range channel {
		go func() {
			client, err := rpc.DialHTTP("tcp", endpointAddr)
			if err != nil {
				log.Printf("rpc.Dial failed: %v", err)
				return
			}
			defer client.Close()

			rpcCall(client)
		}()
	}
}
