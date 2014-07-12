package forwarder

import "fmt"
import "log"
import "net"
import "net/http"
import "net/rpc"

type Forwarder struct{}

type Args struct {
	Addr *net.UDPAddr
	Buf  []byte
}

func (*Forwarder) Send(args *Args, _ *struct{}) error {
	conn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := conn.WriteTo(args.Buf, args.Addr)
	if err != nil {
		return err
	}
	if n != len(args.Buf) {
		return fmt.Errorf("PacketConn.WriteTo failed: short write: %v", n)
	}
	return nil
}

func (*Forwarder) SendAndReply(args *Args, reply *[]byte) error {
	conn, err := net.ListenPacket("udp", "0.0.0.0:0")
	if err != nil {
		return err
	}
	defer conn.Close()

	n, err := conn.WriteTo(args.Buf, args.Addr)
	if err != nil {
		return err
	}
	if n != len(args.Buf) {
		return fmt.Errorf("PacketConn.WriteTo failed: short write: %v", n)
	}

	buf := make([]byte, 1024)
	n, _, err = conn.ReadFrom(buf)
	if err != nil {
		return err
	}

	*reply = buf[:n]
	return nil
}

func Listen(port int) {
	log.Printf("starting forwarding service on port %v", port)

	forwarder := &Forwarder{}
	rpc.Register(forwarder)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		log.Fatal("net.Listen failed: ", err)
	}
	http.Serve(listener, nil)
}
