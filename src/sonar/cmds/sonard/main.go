package main

import (
	"log"
	"net"

	"sonar/ping"
)

func main() {
	ip := net.IPAddr{
		IP: net.ParseIP("8.8.8.8"),
	}

	r, err := ping.Measure(&ip, 10)
	if err != nil {
		log.Panic(err)
	}

	for _, t := range r.Data {
		log.Println(t)
	}
}
