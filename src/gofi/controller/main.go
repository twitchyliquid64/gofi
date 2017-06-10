package main

import (
	"fmt"
	"gofi/udpserv"
	"log"
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

	serv, err := udpserv.New()
	if err != nil {
		log.Printf("Failed to start UDP server: %s", err)
		os.Exit(1)
	}
	defer serv.Close()

	for {
		discoveryPkt := serv.Read()
		discoveryPkt.Debug()
	}
}
