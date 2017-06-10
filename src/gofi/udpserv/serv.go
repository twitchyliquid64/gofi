package udpserv

import (
	"gofi/packet"
	"log"
	"net"
)

// Serv represents a running UDP server
type Serv struct {
	listenAddr *net.UDPAddr
	serverConn *net.UDPConn
	close      chan bool

	output chan *packet.Discovery
}

// New creates a new UDP server bound to the Ubiquiti discovery port.
func New() (*Serv, error) {
	out := Serv{
		close:  make(chan bool),
		output: make(chan *packet.Discovery, 1),
	}
	var err error

	out.listenAddr, err = net.ResolveUDPAddr("udp", ":10001")
	if err != nil {
		return nil, err
	}

	/* Now listen at selected port */
	out.serverConn, err = net.ListenUDP("udp", out.listenAddr)
	if err != nil {
		return nil, err
	}

	go out.mainloop()
	return &out, nil
}

// Close shuts down the server
func (s *Serv) Close() error {
	socketErr := s.serverConn.Close()
	if socketErr != nil {
		return socketErr
	}
	close(s.close)
	return nil
}

func (s *Serv) Read() *packet.Discovery {
	return <-s.output
}

// Recieve waits for the next packet and decodes it.
func (s *Serv) Recieve() (*packet.Discovery, error) {
	buf := make([]byte, 8192)

	n, _, err := s.serverConn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	return packet.DiscoveryDecode(buf[:n])
}

func (s *Serv) mainloop() {
	for {
		select {
		case <-s.close:
			break
		default:
			pkt, err := s.Recieve()
			if err != nil {
				log.Printf("Error reading Discovery packet: %s\n", err)
				break
			}
			s.output <- pkt
		}
	}
}
