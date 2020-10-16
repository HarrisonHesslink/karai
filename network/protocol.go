package network

import (
	// "encoding/hex"
	"log"
	"net"
	"strconv"
	"fmt"
	//"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/db"
	"github.com/scottjg/upnp"
	"github.com/karai/go-karai/peer_manager"
    "time"

)

func Protocol_Init(c *config.Config) {
	var s Server
	var d db.Database
	var p Protocol
	var pmm peer_manager.PeerManager
	upnpMan := new(upnp.Upnp)

	s.u = upnpMan

	pmm.Cf = c
	d.Cf = c
	s.cf = c

	p.dat = &d

	s.prtl = &p
	s.PeerManager =  &pmm

	d.CreateRoot()
	log.Println(s.cf.Lport)
	if !s.GrabExternalIPAddr() {
		log.Panic("unable to grab external ip")
	}

	if !s.AddPortMapping() {
		log.Panic("unable to add port mapping")
	}

	nodeAddress = s.ExternalIP + ":" + strconv.Itoa(s.cf.Lport)
	ln, err := net.Listen("tcp", nodeAddress)
	defer ln.Close()
	if err != nil {
		log.Panic(err)
	}

	// if nodeAddress != KnownNodes[0] {
	// 	s.SendVersion(KnownNodes[0])
	// 	s.SendAddr(KnownNodes[0])
	// }
	// for {
	// 	conn, err := ln.Accept()
	// 	if err != nil {
	// 		log.Panic(err)
	// 	}
	// 	go s.HandleConnection(conn)
	// }
}

// Obtain public ip address
func (s Server) GrabExternalIPAddr() bool {
	upnpMan := *s.u
	err := upnpMan.ExternalIPAddr()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		log.Println("Extnernal" + upnpMan.GatewayOutsideIP)
		 s.ExternalIP = upnpMan.GatewayOutsideIP
		 return true
	}
	return false
}

// Add a port mapping
func (s Server) AddPortMapping() bool {
	upnpMan := *s.u
	if err := upnpMan.AddPortMapping(int(55789), int(55789), int(3600), "TCP", "karai-xeq"); err == nil {
		s.ExternalPort = s.cf.Lport
		return true
	} else {
		fmt.Println("Port mapping failed.")
		return false
	}
	return false
}

func (s Server) UpdatePortMapping() {

	//update port mapping every 30 minutes

	for {
		 _ = s.AddPortMapping()




		 time.Sleep(30 * time.Minute)
	}
}


func (s Server) addPeer(addr string) {
	s.Peers = append(s.Peers, addr)
}
