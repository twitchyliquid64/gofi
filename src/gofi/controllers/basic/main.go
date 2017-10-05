package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gofi/manager"
	"gofi/packet"
	"log"
	"net/http"
	"os"
)

var ssid = flag.String("ssid", "gofi", "Network name")
var password = flag.String("pw", "fiog", "Network password")
var do5G = flag.Bool("enable_5g", true, "Make network available on 5G as well as 2.4G")
var bandSteer = flag.Bool("enable_bandsteering", false, "Steer clients to 5G network")
var txPower = flag.Int("tx", 0, "(optional) TX power in DB, defaults to auto")
var minRSSI = flag.Int("min_rssi", 0, "(optional) Station RSSI at which it is deauthed, defaults to disabled")
var localAddress = flag.String("addr", "", "(optional) Controller LAN IP - autodetected if not set")
var configPath = flag.String("statefile", "", "Path to location to store state")
var infoServer = flag.String("infoserv", "", "Address to host the infoserv at. Infoserv disabled if not provided.")

var lastInformForMAC map[string]*packet.InformData

func main() {
	lastInformForMAC = map[string]*packet.InformData{}
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
	informChan := make(chan *packet.InformData, 5)
	go func() {
		for i := range informChan {
			lastInformForMAC[i.Mac] = i
		}
	}()

	if *infoServer != "" {
		fmt.Println("Infoserver will run on", *infoServer)
		h := http.NewServeMux()
		h.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			e := json.NewEncoder(rw)
			e.Encode(lastInformForMAC)
		})
		go func() {
			fmt.Println(http.ListenAndServe(*infoServer, h))
		}()
	}

	manager, err := manager.New(":8421", controllerAddr, nil, onDiscoveryPacket, onControllerDoesntKnowAP, informChan)
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
