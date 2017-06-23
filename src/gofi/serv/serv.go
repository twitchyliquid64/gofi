package serv

// Compartmentizes network request/response logic

import (
	"fmt"
	"gofi/packet"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

type informHandler interface {
	HandleInform(string, *packet.Inform) ([]byte, error)
}

// Serv represents a running server
type Serv struct {
	listenAddr    *net.UDPAddr
	serverConn    *net.UDPConn
	informHandler informHandler

	DiscoveryPackets chan *packet.Discovery

	httpServ *http.Server

	close chan bool
}

// New creates a new server bound to the Ubiquiti discovery port and the given port for the HTTP server.
func New(ihandler informHandler, httpListener string) (*Serv, error) {
	out := Serv{
		close:            make(chan bool),
		DiscoveryPackets: make(chan *packet.Discovery, 1),
		informHandler:    ihandler,
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

	out.makeHTTPServer(httpListener)

	go out.httpMainloop()
	go out.discoveryMainloop()
	return &out, nil
}

func (s *Serv) makeHTTPServer(listener string) {
	s.httpServ = &http.Server{Addr: listener}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello world\n")
	})
	http.HandleFunc("/inform", func(w http.ResponseWriter, r *http.Request) {
		informPkt, err := packet.InformDecode(r.Body)
		if err != nil {
			fmt.Println("Error decoding Inform: ", err)
		} else {
			data, err := s.informHandler.HandleInform(r.RemoteAddr, informPkt)
			if err != nil {
				fmt.Printf("HandleInform() err: %s\n", err)
				return
			}
			w.Header().Set("Content-Type", "application/x-binary")
			w.Header().Set("Content-Length", strconv.Itoa(len(data)))
			w.Header().Set("User-Agent", "Unifi Controller")
			w.Header().Set("Connection", "close")
			w.Write(data)
		}
	})
}

// Close shuts down the server
func (s *Serv) Close() error {
	socketErr := s.serverConn.Close()
	httpErr := s.httpServ.Close()
	close(s.close)
	if socketErr != nil {
		return socketErr
	}
	return httpErr
}

// Recieve waits for the next packet and decodes it.
func (s *Serv) Recieve() (*packet.Discovery, error) {
	buf := make([]byte, 8192)

	n, addr, err := s.serverConn.ReadFromUDP(buf)
	if err != nil {
		return nil, err
	}
	return packet.DiscoveryDecode(addr, buf[:n])
}

func (s *Serv) httpMainloop() {
	err := s.httpServ.ListenAndServe()
	fmt.Printf("HTTP Server error: %s\n", err)
}

func (s *Serv) discoveryMainloop() {
	defer close(s.DiscoveryPackets)
	for {
		select {
		case <-s.close:
			return
		default:
			pkt, err := s.Recieve()
			if err != nil {
				log.Printf("Error reading Discovery packet: %s\n", err)
				return
			}
			s.DiscoveryPackets <- pkt
		}
	}
}
