package main

import (
	"fmt"
	"gofi/config"
	"gofi/manager"
	"log"
	"os"
)

func main() {
	laddr, err := localAddr()
	if err != nil {
		log.Println("Error fetching local address:", err)
		return
	}
	log.Printf("Started, controller running on %s\n", laddr)

	manager, err := manager.New(":8421", laddr.String(), &config.Config{
		Networks: []config.Network{
			config.Network{
				SSID: "kek",
				Pass: "the_shrekkening",
			},
			config.Network{
				SSID:   "kek",
				Pass:   "the_shrekkening",
				Is5Ghz: true,
			},
		},
		Bandsteer: config.SteerSettings{
			Enabled: true,
			Mode:    config.SteerPrefer5G,
		},
	}, nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer manager.Close()
	fmt.Println(manager.Run())
}
