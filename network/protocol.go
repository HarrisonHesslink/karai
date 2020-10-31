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
	var d database.Database
	var p Protocol
	var peer_list PeerList
	var sync Syncer
	var mempool MemPool
	sync.Connected = false
	sync.Synced = false
	p.Mempool = &mempool

	p.Sync = &sync
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
				if s.Prtl.Sync.Connected == true {
					req, err := ioutil.ReadAll(ctx.Body)
					if err != nil {
						log.Panic(err)
					}
					go s.HandleConnection(req, ctx)
				}
			},
		},
	}

	defer s.node.Shutdown()

	err = s.node.Start(s.ExternalIP)
	s.node.Probe("167.172.156.118:4201")
	s.node.Probe("157.230.91.2:4201")
	s.node.Probe(":4201")
	s.Prtl.Sync.Connected = true
	


	if err != nil {
		log.Println("Unable to connect")
	}

	go s.LookForNodes()

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
		if s.pl.Count < 9 {
			new_ids := s.node.Bootstrap()

			//probe new nodes

			for _, peer := range new_ids {
				s.node.Probe(peer.Host.String() + ":" + strconv.Itoa(int(peer.Port)))
			}

			providers := s.node.ProvidersFor("karai-xeq")
			//log.Println(strconv.Itoa(len(providers)))
			for _, provider := range providers {
					go s.SendVersion(provider)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func (s *Server) NewDataTxFromCore(req transaction.Request_Oracle_Data) {

	if s.Prtl.MyNodeKey == "" {
		s.Prtl.MyNodeKey = req.PubKey
	}
	if !s.inMempool(req.Hash) {
		s.Prtl.Mempool.Transactions = append(s.Prtl.Mempool.Transactions, req)
	}

	go s.BroadCastOracleData(req)
}

func (s *Server) NewConsensusTXFromCore(req transaction.Request_Consensus) {
	req_string, _ := json.Marshal(req)

	if s.Prtl.MyNodeKey == "" {
		s.Prtl.MyNodeKey = req.PubKey
	}

	s.Prtl.ConsensusNode = req.PubKey

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	new_tx := transaction.CreateTransaction("1", txPrev, req_string, []string{}, []string{})
	if !s.Prtl.Dat.HaveTx(new_tx.Hash) {
		go s.Prtl.Dat.CommitDBTx(new_tx)
		go s.BroadCastTX(new_tx)
	}
}

func (s *Server) CreateContract(asset string, denom string) {
	var txPrev string
	contract := transaction.Request_Contract{asset, denom}
	json_contract,_ := json.Marshal(contract)

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	tx := transaction.CreateTransaction("3", txPrev, []byte(json_contract), []string{}, []string{})

	if !s.Prtl.Dat.HaveTx(tx.Hash) {
		go s.Prtl.Dat.CommitDBTx(tx) 
		go s.BroadCastTX(tx)
	}
	log.Println("Created Contract " + tx.Hash[:8]+ ": " + asset + "/" + denom)
}

/*

CheckNode checks if a node should be able to put data on the contract takes a Transaction

*/
func (s *Server) CheckNode(tx transaction.Transaction) bool {

	checks_out := false
	var hash string
	var tx_data string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash, tx_data FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' && tx_epoc=$1 ORDER BY tx_time DESC", tx.Epoc).Scan(&hash, &tx_data)

	if hash != "" {
		checks_out = true
	}

	var last_consensus transaction.Request_Consensus
	err := json.Unmarshal([]byte(tx_data), &last_consensus)
	if err != nil {
		//unable to parse last consensus ? this should never happen
		log.Println("Failed to Parse Last Consensus TX on Cehck")
		return false
	}

	//get interface for checks [Request_Consensus, Request_Oracle_Data, Request_Contract]

	result := tx.ParseInterface()
	if result == nil {
		return false
	}

	switch v := result.(type) {
	case transaction.Request_Consensus:
		isFound := false
		for _, key := range last_consensus.Data {
			if key == v.PubKey {
				isFound = true
				break
			}
		}

		if !isFound {
			return false
		}

		


		// here v has type T
		break;
	case transaction.Request_Oracle_Data:
		// here v has type S
		break;
	case transaction.Request_Contract:
		break;
	default:
		return false;
	}

	return checks_out
}

func (s *Server) GetContractMap() map[string]string {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var Contracts map[string]string
	Contracts = make(map[string]string)

	//loop through to find oracle data
	rows, err := db.Queryx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_type='3' ORDER BY tx_time DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var this_tx transaction.Transaction
		err = rows.StructScan(&this_tx)
		if err != nil {
			// handle this error
			log.Panic(err)
		}
		var data_prev string
		_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_epoc=$1 ORDER BY tx_time DESC", this_tx.Hash).Scan(&data_prev)
		Contracts[this_tx.Hash] = data_prev
	}
	err = rows.Err()
	if err != nil {
		log.Panic(err)
	}

	return Contracts
}
func (s *Server) CreateTrustedData(block_height string) {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	contract_data_map := s.sortOracleDataMap(block_height)
	
	filtered_data_map, trusted_data_map := filterOracleDataMap(contract_data_map)

	log.Println("Creating Trust Data TX")

	for _, contract_array := range filtered_data_map {
		var prev string
		_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_epoc=$1 ORDER BY tx_time DESC", contract_array[0].Epoc).Scan(&prev)

		if prev == "" {
			return
		}


		trusted_data := transaction.Trusted_Data{contract_array, trusted_data_map[contract_array[0].Epoc]}

		new_tx := transaction.CreateTrustedTransaction(prev, trusted_data)

		go s.Prtl.Dat.CommitDBTx(new_tx) 		
		go s.BroadCastTX(new_tx)
	}
}


