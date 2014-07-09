package forwarder

import "fmt"
import "log"
import "net"
import "net/http"
import "net/rpc"

type Forwarder struct{}

type Args struct {
	Addr *net.UDPAddr
	Req  []byte
}

func (t *Forwarder) Forward(args *Args, resp *[]byte) error {
	log.Printf("Handling a forwarding request...")

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := conn.WriteTo(args.Req, args.Addr)
	if err != nil {
		return err
	}
	if n != len(args.Req) {
		return fmt.Errorf("PacketConn.WriteTo failed: short write: %v", n)
	}

	buf := make([]byte, 1024)
	// FIXME: Unclear if the source IP needs to be forged when rebroadcast.
	n, _, err = conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	log.Printf("Responding with %v...", buf[:n])
	*resp = buf[:n]
	return nil
}

func Listen() {
	forwarder := &Forwarder{}
	rpc.Register(forwarder)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", ":8714")
	if err != nil {
		log.Fatal("net.Listen failed: ", err)
	}
	http.Serve(listener, nil)
}
