package network

import (
	// "encoding/hex"
	"log"
	//"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/db"
	"github.com/karai/go-karai/peer_manager"
	"github.com/harrisonhesslink/flatend"
	//"strconv"
	"github.com/glendc/go-external-ip"
	//"github.com/karai/go-karai/transaction"
	//"io/ioutil"
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
	s.node = &flatend.Node{
		PublicAddr: ":4201",
		BindAddrs:  []string{":4201"},
		SecretKey:  flatend.GenerateSecretKey(),
		Services: map[string]flatend.Handler{
			"karai-xeq": func(ctx *flatend.Context) {

				// req, err := ioutil.ReadAll(ctx.Body)
				// if err != nil {
				// 	log.Panic(err)
				// }
			
				//log.Println(string(req))
				s.HandleConnection(ctx)
			},
		},
	}
	defer s.node.Shutdown()

	err = s.node.Start(s.ExternalIP)

	s.node.Probe("157.230.91.2:4201")

	if err != nil {
		log.Println("Unable to connect")
		log.Panic(err)
	}

	s.SendVersion()

	

	// for _, provider := range providers {
	// 	_, err := provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
	// 	if err != nil {
	// 		//fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
	// 	}
	// }

	select {}
}

func (s Server) addPeer(addr string) {
	s.Peers = append(s.Peers, addr)
}
