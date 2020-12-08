package main

import (
	config "github.com/harrisonhesslink/pythia/configuration"
	"github.com/harrisonhesslink/pythia/network"
)

// Pythia Main
func main() {
	c := config.Config_Init()
	flags(&c)

	var s network.Server
	go inputHandler(&s)
	network.ProtocolInit(&c, &s)

}
