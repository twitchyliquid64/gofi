package main

import (
	"flag"
	"fmt"
	"gofi/config"
	"gofi/manager"
	"log"
	"os"
)

var ssid = flag.String("ssid", "gofi", "Network name")
var password = flag.String("pw", "fiog", "Network password")
var do5G = flag.Bool("enable_5g", true, "Make network available on 5G as well as 2.4G")
var bandSteer = flag.Bool("enable_bandsteering", false, "Steer clients to 5G network")
var txPower = flag.Int("tx", 0, "(optional) TX power in DB, defaults to auto")
var minRSSI = flag.Int("min_rssi", 0, "(optional) Station RSSI at which it is deauthed, defaults to disabled")
var localAddress = flag.String("addr", "", "Controller LAN IP - autodetected if not set")

func main() {
	flag.Parse()
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
	c := &config.Config{
		Networks: []config.Network{
			config.Network{
				SSID: *ssid,
				Pass: *password,
			},
		},
		Bandsteer: config.SteerSettings{
			Enabled: *bandSteer,
			Mode:    config.SteerPrefer5G,
		},
		Txpower: *txPower,
		MinRSSI: *minRSSI,
	}

	if *do5G {
		c.Networks = append(c.Networks, config.Network{
			SSID:   *ssid,
			Pass:   *password,
			Is5Ghz: true,
		})
	}

	manager, err := manager.New(":8421", controllerAddr, c, nil, nil, nil)
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
