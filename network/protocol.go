package network

import (
	"fmt"
	"log"

	api "github.com/harrisonhesslink/pythia/api"
	config "github.com/harrisonhesslink/pythia/configuration"
	contract "github.com/harrisonhesslink/pythia/contract"
	"github.com/harrisonhesslink/pythia/database"

	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	externalip "github.com/glendc/go-external-ip"
	"github.com/harrisonhesslink/flatend"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
	"github.com/lithdew/kademlia"

	_ "github.com/lib/pq"
	//"github.com/gorilla/websocket"
)

/*

ProtocolInit = init all of the protocol

*/
func ProtocolInit(c *config.Config, s *Server) {

	var p Protocol
	var peer_list PeerList
	var sync Syncer
	sync.Connected = false
	sync.Synced = false

	p.Sync = &sync
	s.pl = &peer_list
	s.cf = c

	p.Dat = database.NewDataBase(c)
	p.Mempool = NewMemPool()

	s.Prtl = &p

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
	s.Prtl.Sync.Connected = true

	if err != nil {
		log.Println("Unable to connect")
	}

	go s.LookForNodes()

	select {}
}

/*

HandleCall = Handle a call from p2p

*/
func (s *Server) HandleCall(stream *flatend.Stream) {
	req, err := ioutil.ReadAll(stream.Reader)
	if err != nil {
		log.Panic(err)
	}
	go s.HandleConnection(req, nil)
}

/*

GetProviderFromID = Get provider from id

*/
func (s *Server) GetProviderFromID(id *kademlia.ID) *flatend.Provider {
	providers := s.node.ProvidersFor("karai-xeq")
	if providers != nil && len(providers) > 0 {
		for _, provider := range providers {
			if provider.GetID().Pub.String() == id.Pub.String() {
				return provider
			}
		}
	}
	return nil
}

/*

LookForNodes = Look for peers not known

*/
func (s *Server) LookForNodes() {
	for {
		if s.pl.Count < 9 {

			providers := s.node.ProvidersFor("karai-xeq")
			for _, provider := range providers {
				go s.SendVersion(provider)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

//NewDataTxFromCore = Go through all contracts and send data out
func (s *Server) NewDataTxFromCore(req transaction.NewBlock) {
	log.Println(req.Pubkey)
	if s.Prtl.MyNodeKey == "" {
		s.Prtl.MyNodeKey = req.Pubkey
	}

	var agg float64

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	rows, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='3' ORDER BY tx_time DESC")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var this_tx transaction.Transaction
		err = rows.StructScan(&this_tx)
		if err != nil {
			// handle this error
			log.Panic(err)
		}

		var contract contract.Contract
		err := json.Unmarshal([]byte(this_tx.Data), &contract)
		if err != nil {
			log.Panic(err)
		}

		data, r := api.MakeRequest(contract)
		if r {
			for _, v := range data {
				f, _ := strconv.ParseFloat(v, 64)
				agg += f
			}
			agg = agg / float64(len(data))

			var oracledata transaction.OracleData
			oracledata.Height = req.Height
			oracledata.Pubkey = req.Pubkey
			oracledata.Price = agg
			oracledata.Contract = this_tx.Hash
			oracledata.Hash = ""
			oracledata.Signature = ""
			sig, hash := api.CoreSign(oracledata)

			oracledata.Hash = hash
			oracledata.Signature = sig

			s.BroadCastOracleData(oracledata)
		}

	}
	err = rows.Err()
	if err != nil {
		log.Panic(err)
	}
	rows.Close()

	// if s.Prtl.Mempool.addOracleData() {
	// 	go s.BroadCastOracleData(req)
	// }
}

//NewConsensusTXFromCore = create v1 tx
func (s *Server) NewConsensusTXFromCore(req transaction.NewBlock) {
	req_string, _ := json.Marshal(req)

	if s.Prtl.MyNodeKey == "" {
		s.Prtl.MyNodeKey = req.Pubkey
	}
	log.Println(req.Pubkey)
	s.Prtl.ConsensusNode = req.Pubkey

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	new_tx := transaction.CreateTransaction("1", txPrev, req_string, []string{}, []string{})
	if !s.Prtl.Dat.HaveTx(new_tx.Hash) {
		s.Prtl.Dat.CommitDBTx(new_tx)
		go s.BroadCastTX(new_tx)
	}
}

//CreateContract make new contract uploaded fron config.json
func (s *Server) CreateContract() {
	var txPrev string
	file, _ := ioutil.ReadFile("contract.json")

	data := contract.Contract{}

	_ = json.Unmarshal([]byte(file), &data)

	jsonContract, _ := json.Marshal(data)

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	tx := transaction.CreateTransaction("3", txPrev, []byte(jsonContract), []string{}, []string{})

	if !s.Prtl.Dat.HaveTx(tx.Hash) {
		s.Prtl.Dat.CommitDBTx(tx)
		go s.BroadCastTX(tx)
	}
	log.Println("Created Contract " + tx.Hash[:8])
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

	_ = db.QueryRow("SELECT tx_hash, tx_data FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_type='1' && tx_epoc=$1 ORDER BY tx_time DESC", tx.Epoc).Scan(&hash, &tx_data)

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
		break
	case transaction.Request_Oracle_Data:
		// here v has type S
		break
	case transaction.Request_Contract:
		break
	default:
		return false
	}

	return checks_out
}

/*

GetContractMap creates contract map and their last known tx
*/
func (s *Server) GetContractMap() map[string]string {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var Contracts map[string]string
	Contracts = make(map[string]string)

	//loop through to find oracle data
	rows, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='3' ORDER BY tx_time DESC")
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
		_ = db.QueryRow("SELECT tx_hash FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", this_tx.Hash).Scan(&data_prev)
		Contracts[this_tx.Hash] = data_prev
	}
	err = rows.Err()
	if err != nil {
		log.Panic(err)
	}

	return Contracts
}

/*

CreateTrustedData creates trusted data source from all known tx
*/
func (s *Server) CreateTrustedData(block_height string) {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	contract_data_map := s.Prtl.Mempool.SortOracleDataMap(block_height)

	filtered_data_map, trusted_data_map := FilterOracleDataMap(contract_data_map)

	log.Println("Creating Trust Data TX for block: " + block_height)
	log.Println("Size of filtered_data_map: " + strconv.Itoa(len(filtered_data_map)))

	for _, contract_array := range filtered_data_map {
		log.Println("Size of contract_array: " + strconv.Itoa(len(contract_array)))
		if len(contract_array) > 1 {

			var lastTrustedTx transaction.Transaction
			_ = db.QueryRowx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", contract_array[0].Contract).StructScan(&lastTrustedTx)
			prev := lastTrustedTx.Hash
			if prev == "" {
				return
			}

			multi := 1.0

			var contractTx transaction.Transaction
			_ = db.QueryRowx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_hash=$1 ORDER BY tx_time DESC", contract_array[0].Contract).StructScan(&contractTx)

			contract := contract.Contract{}
			json.Unmarshal([]byte(contractTx.Data), &contract)

			if contract.ContractRef != "" {
				var lastContractRef transaction.Transaction
				_ = db.QueryRowx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", contract.ContractRef).StructScan(&lastContractRef)

				if lastContractRef.Hash == "" {
					log.Println("Unable to query last contract ref!")
				}

				td := transaction.Trusted_Data{}
				json.Unmarshal([]byte(lastContractRef.Data), &td)
				if td.TrustedAnswer > 0 {
					multi = td.TrustedAnswer

				} else {
					multi = 1.0
				}

			} else {
				multi = 1.0
			}

			price := trusted_data_map[contract_array[0].Contract] * multi
			fmt.Print("Trusted Answer: ")
			fmt.Println(price)
			send := false
			if contract.Threshold != "" {
				if s, err := strconv.ParseFloat(contract.Threshold, 64); err != nil {
					log.Println(s)
					if s >= 0.0 {
						ltd := transaction.Trusted_Data{}
						json.Unmarshal([]byte(lastTrustedTx.Data), &ltd)

						change := PercentageChange(ltd.TrustedAnswer, price)

						if change >= s {
							send = true
						}
					}
				}
			} else {
				send = true
			}

			height, _ := strconv.Atoi(block_height)

			go s.Prtl.Mempool.PruneHeight(strconv.Itoa(height - 5))

			if send {
				trusted_data := transaction.Trusted_Data{contract_array, price}

				new_tx := transaction.CreateTrustedTransaction(prev, trusted_data)

				s.Prtl.Dat.CommitDBTx(new_tx)
				go s.BroadCastTX(new_tx)
			}
		}
	}
}
