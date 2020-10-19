package network

import (
	// "encoding/hex"
	"log"
	//"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/db"
	"github.com/harrisonhesslink/flatend"
	"strconv"
	"github.com/glendc/go-external-ip"
	//"github.com/karai/go-karai/transaction"
	"io/ioutil"
	"time"
	"github.com/lithdew/kademlia"
	"encoding/json"

)

func Protocol_Init(c *config.Config, s *Server) {
	var d db.Database
	var p Protocol
	var peer_list PeerList

	s.pl = &peer_list
	d.Cf = c
	s.cf = c

	p.Dat = &d

	s.Prtl = &p
	d.DB_init()

	go s.RestAPI()

	log.Println(c.Lport)

  	consensus := externalip.DefaultConsensus(nil, nil)
    // Get your IP,
    // which is never <nil> when err is <nil>.
    ip, err := consensus.ExternalIP()
    if err != nil {
		log.Panic(ip)
	}
	s.ExternalIP = ip.String()
	s.node = &flatend.Node{
		PublicAddr: ":" + strconv.Itoa(c.Lport),
		BindAddrs:  []string{":" + strconv.Itoa(c.Lport)},
		SecretKey:  flatend.GenerateSecretKey(),
		Services: map[string]flatend.Handler{
			"karai-xeq": func(ctx *flatend.Context) {

				req, err := ioutil.ReadAll(ctx.Body)
				if err != nil {
					log.Panic(err)
				}
			
				s.HandleConnection(req, ctx)

				ctx.Write([]byte("close"))
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

	go s.LookForNodes()

	log.Println("Active Peer Count with streams: " + strconv.Itoa(s.pl.Count))

	// for _, provider := range providers {
	// 	_, err := provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
	// 	if err != nil {
	// 		//fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
	// 	}
	// }

	select {}
}

func (s *Server) HandleCall(stream *flatend.Stream) {
	req, err := ioutil.ReadAll(stream.Reader)
	if err != nil {
		log.Panic(err)
	}
	go s.HandleConnection(req, nil)
}

func (s *Server) GetProviderFromID(id  *kademlia.ID) *flatend.Provider {
	providers := s.node.ProvidersFor("karai-xeq")
	for _, provider := range providers {
		if provider.GetID().Pub.String() == id.Pub.String(){
			return provider
		}
	}
	return nil
}

func (s *Server) LookForNodes() {
	for {
		new_ids := s.node.Bootstrap()

		//probe new nodes

		for _, peer := range new_ids {
			log.Println(peer.Host.String() + ":" + strconv.Itoa(int(peer.Port)))
			s.node.Probe(peer.Host.String() + ":" + strconv.Itoa(int(peer.Port)))
		}



		providers := s.node.ProvidersFor("karai-xeq")
		log.Println(strconv.Itoa(len(providers)))
		for _, provider := range providers {
	
				stream := s.SendVersion(provider)
				if s.pl.Count < 9 {
					s.pl.Peers = append(s.pl.Peers, Peer{provider.GetID(), provider})
					s.pl.Count++
				}
	
				s.HandleCall(stream)
		}

		time.Sleep(1 * time.Minute)
	}
}

func (s *Server) NewDataTxFromCore(req Request) {
	req_string, _ := json.Marshal(req)
	log.Println(string(req_string))
}