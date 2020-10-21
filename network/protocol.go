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
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"io/ioutil"
	"time"
	"github.com/lithdew/kademlia"
	"encoding/json"
	// "github.com/gorilla/websocket"

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
			
				go s.HandleConnection(req, ctx)

				ctx.Write([]byte("close"))
			},
		},
	}
	defer s.node.Shutdown()

	err = s.node.Start(s.ExternalIP)

	s.node.Probe("167.172.156.118:4201")

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

func (s *Server) NewDataTxFromCore(req transaction.Request_Data_TX) {
	req_string, _ := json.Marshal(req)

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var txData string
	i := 0
	for i <= 10 {
		_ = db.QueryRow("SELECT tx_data FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txData)


		last_consensus_data := transaction.Request_Consensus_TX{}
		err := json.Unmarshal([]byte(txData), &last_consensus_data)
		if err != nil {
			log.Println("Unable to parse tx_data")
			continue;
		}

		if last_consensus_data.Height == req.Height {
			break;
		}

		i++
		time.Sleep(5 * time.Second)
	}

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	var txhash_on_epoc []string
	var txdata_on_epoc []string

	//Grab all first txes on epoc 
	rows, err := db.Query("SELECT tx_hash, tx_data FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_epoc=$1 ORDER BY tx_time DESC" , txPrev)
	if err != nil {
		// handle this error better than this
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var tx_hash string
		var tx_data string
		err = rows.Scan(&tx_hash, &tx_data)
		if err != nil {
			// handle this error
			log.Panic(err)
		}
		
		txhash_on_epoc = append(txhash_on_epoc, tx_hash)
		txdata_on_epoc = append(txdata_on_epoc, tx_data)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		log.Panic(err)
	}

	new_tx := transaction.CreateTransaction("2", txPrev, req_string, txhash_on_epoc, txdata_on_epoc)

	s.Prtl.Dat.CommitDBTx(new_tx)

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
	
	s.BroadCastTX(new_tx)
}

// func (s *Server) HandleAPISocket(c *websocket.Conn) {
// 	for {
// 		mt, message, err := c.ReadMessage()
// 		if err != nil {
// 			log.Println("read:", err)
// 			break
// 		}
// 		log.Printf("recv: %s", message)
// 		err = c.WriteMessage(mt, message)
// 		if err != nil {
// 			log.Println("write:", err)
// 			break
// 		}
// 	}
// }

