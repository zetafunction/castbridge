package forwarder

import (
	"log"
	"net/rpc"
)

type Request struct {
	*Args
	OnReply func([]byte)
}

func NewClient(endpointAddr string) chan<- *Request {
	channel := make(chan *Request, 100)
	go handleRequests(endpointAddr, channel)
	return channel
}

func handleRequests(endpointAddr string, channel <-chan *Request) {
	for req := range channel {
		go func() {
			client, err := rpc.DialHTTP("tcp", endpointAddr+":8714")
			if err != nil {
				log.Printf("rpc.Dial failed: %v", err)
				return
			}
			defer client.Close()

			var reply *[]byte
			if err := client.Call("Fowarder.Forward", req, reply); err != nil {
				log.Printf("Client.Call failed: %v", err)
				return
			}
			req.OnReply(*reply)
		}()
	}
}
