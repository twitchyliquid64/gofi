package main

import (
	"flag"
	"fmt"
	"gofi/manager"
	"log"
	"os"
)

var ssid = flag.String("ssid", "gofi", "Network name")
var password = flag.String("pw", "fiog", "Network password")
var do5G = flag.Bool("enable_5g", true, "Make network available on 5G as well as 2.4G")
var bandSteer = flag.Bool("enable_bandsteering", false, "Steer clients to 5G network")
var localAddress = flag.String("addr", "", "(optional) Controller LAN IP - autodetected if not set")
var configPath = flag.String("statefile", "", "Path to location to store state")

func main() {
	flag.Parse()

	errLoad := loadConfig(*configPath)
	if errLoad != nil {
		fmt.Println("Error:", errLoad)
		os.Exit(1)
	}

	if *bandSteer && !*do5G {
		fmt.Println("Error: Cannot bandsteer without 5G networks enabled")
		os.Exit(1)
	}

	controllerAddr := *localAddress
	if controllerAddr == "" {
		laddr, err := localAddr()
		if err != nil {
			log.Println("Error fetching local address:", err)
			os.Exit(1)
		}
		controllerAddr = laddr.String()
	}

	log.Printf("Controller will run on %s\n", controllerAddr)

	manager, err := manager.New(":8421", controllerAddr, nil, onDiscoveryPacket, onControllerDoesntKnowAP)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer manager.Close()
	err = manager.Run()
	if err != nil {
		fmt.Println("Error starting manager: ", err)
	}
}
