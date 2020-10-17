package network

import (
	// "encoding/hex"
	"log"
	//"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/db"
	"github.com/karai/go-karai/peer_manager"
	"github.com/lithdew/flatend"
	//"strconv"
	"github.com/glendc/go-external-ip"

)

func Protocol_Init(c *config.Config) {
	var s Server
	var d db.Database
	var p Protocol
	var pmm peer_manager.PeerManager

	pmm.Cf = c
	d.Cf = c
	s.cf = c

	p.dat = &d

	s.prtl = &p
	s.PeerManager =  &pmm
	d.DB_init()


  	consensus := externalip.DefaultConsensus(nil, nil)
    // Get your IP,
    // which is never <nil> when err is <nil>.
    ip, err := consensus.ExternalIP()
    if err != nil {
		log.Panic(ip)
	}
	s.ExternalIP = ip.String()
	log.Println(s.ExternalIP)
	s.node = &flatend.Node{
		PublicAddr: "0.0.0.0:4201",
		SecretKey:  flatend.GenerateSecretKey(),
		Services: map[string]flatend.Handler{
			"karai-xeq": func(ctx *flatend.Context) {
				log.Println(ctx.ID.Host.String())
				s.HandleConnection(ctx)
			},
		},
	}
	defer s.node.Shutdown()
	err = s.node.Start("157.230.91.2:4201")
	if err != nil {
		log.Println("Unable to connect")
		log.Panic(err)
	}

	s.SendVersion("")


	select {}
}


func (s Server) addPeer(addr string) {
	s.Peers = append(s.Peers, addr)
}
