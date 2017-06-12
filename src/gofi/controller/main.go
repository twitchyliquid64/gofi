package main

import (
	"errors"
	"fmt"
	"gofi/manager"
	"log"
	"net"
	"os"
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

	manager, err := manager.New(":8421", laddr.String())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer manager.Close()
	fmt.Println(manager.Run())
}
