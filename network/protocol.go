package network

import (
	// "encoding/hex"
	"log"
	//"github.com/karai/go-karai/database"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/database"
	"github.com/harrisonhesslink/flatend"
	"strconv"
	"github.com/glendc/go-external-ip"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"io/ioutil"
	"time"
	"github.com/lithdew/kademlia"
	"encoding/json"
	//"github.com/gorilla/websocket"

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


			
				go s.HandleConnection(req, ctx)
			},
		},
	}
	defer s.node.Shutdown()

	err = s.node.Start(s.ExternalIP)

	go s.node.Probe("167.172.156.118:4201")

	if err != nil {
		log.Println("Unable to connect")
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
		//log.Println(strconv.Itoa(len(providers)))
		for _, provider := range providers {
	
				s.SendVersion(provider)
				if s.pl.Count < 9 {
					s.pl.Peers = append(s.pl.Peers, Peer{provider.GetID(), provider})
					s.pl.Count++
				}	
		}

		time.Sleep(1 * time.Minute)
	}
}

func (s *Server) NewDataTxFromCore(req transaction.Request_Data_TX) {
	req_string, _ := json.Marshal(req)

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='2' AND tx_epoc=$1 ORDER BY tx_time DESC", req.Epoc).Scan(&txPrev)
	
	new_tx := transaction.CreateTransaction("2", txPrev, req_string, []string{}, []string{})

	s.Prtl.Dat.CommitDBTx(new_tx)
	json_tx, _ := json.Marshal(new_tx)

	for _, conn := range s.Sockets {
		if err := conn.WriteMessage(1, json_tx); err != nil {
			log.Println(err)
			return
		}
	}
	s.BroadCastTX(new_tx)
}

func (s *Server) NewConsensusTXFromCore(req transaction.Request_Consensus_TX) {
	req_string, _ := json.Marshal(req)

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	new_tx := transaction.CreateTransaction("1", txPrev, req_string, []string{}, []string{})
		
	s.Prtl.Dat.CommitDBTx(new_tx)
	json_string, _ := json.Marshal(new_tx)
	for _, conn := range s.Sockets {
		if err := conn.WriteMessage(1, json_string); err != nil {
			log.Println(err)
			return
		}
	}
	s.BroadCastTX(new_tx)
}

type Contract struct {
	Asset string`json:asset`
	Denom string`json:denom`
}

func (s *Server) CreateContract(asset string, denom string) {
	var txPrev string
	contract := Contract{asset, denom}
	json_contract,_ := json.Marshal(contract)

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	tx := transaction.CreateTransaction("3", txPrev, []byte(json_contract), []string{}, []string{})
	log.Println("Created Contract " + tx.Hash[:8]+ ": " + asset + "/" + denom)

	if !s.Prtl.Dat.HaveTx(tx.Hash) {
		s.Prtl.Dat.CommitDBTx(tx)

		json_tx,_ := json.Marshal(tx)

		for _, conn := range s.Sockets {
			if err := conn.WriteMessage(1, json_tx); err != nil {
				log.Println(err)
				return
			}
		}
	
		s.BroadCastTX(tx)
	}

}

