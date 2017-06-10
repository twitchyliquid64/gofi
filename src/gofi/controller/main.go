package main

import (
	"fmt"
	"gofi/packet"
	"log"
	"net"
	"os"
)

/* A Simple function to verify error */
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	log.Println("Started")

	/* Lets prepare a address at any address at port 10001*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 8192)

	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Failed ReadFromUDP(): %s\n", err)
			break
		}
		fmt.Printf("Recieved %d bytes from %s\n", n, addr)
		discovery, err := packet.DiscoveryDecode(buf[:n])
		if err != nil {
			fmt.Printf("Error decoding discovery: %s\n", err)
		} else {
			discovery.Debug()
		}
	}
}
