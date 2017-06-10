package main

import (
	"errors"
	"fmt"
	"gofi/adopt"
	"gofi/serv"
	"log"
	"net"
	"os"
	"strings"
)

func localAddr() (net.IP, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			if !ip.IsLoopback() {
				return ip, nil
			}
		}
	}
	return nil, errors.New("Could not find appropriate interface")
}

// CheckError terminates the program if the error is non nil
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	laddr, err := localAddr()
	if err != nil {
		log.Println("Error fetching local address:", err)
		return
	}
	log.Printf("Started, controller running on %s\n", laddr)

	serv, err := serv.New(":8421")
	if err != nil {
		log.Printf("Failed to start UDP server: %s", err)
		os.Exit(1)
	}
	defer serv.Close()

	for {
		discoveryPkt := <-serv.DiscoveryPackets
		discoveryPkt.Debug()
		adoptCfg := adopt.NewConfig(strings.Split(discoveryPkt.IPInfo.String(), ":")[0]+":22", laddr.To4().String()+":8421")
		fmt.Println("adoption configuration:", adoptCfg)
		adoptErr := adopt.Adopt(adoptCfg)
		fmt.Println("adoption operation:", adoptErr)
	}
}
