package network

import (
	"context"
	"fmt"
	"sync"

	api "github.com/harrisonhesslink/pythia/api"
	config "github.com/harrisonhesslink/pythia/configuration"
	contract "github.com/harrisonhesslink/pythia/contract"
	"github.com/harrisonhesslink/pythia/database"

	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/harrisonhesslink/flatend"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
	"github.com/lithdew/kademlia"
	log "github.com/sirupsen/logrus"
)

var (
	GeneralChannel   = "general-channel"
	MiningChannel    = "contract-channel"
	FullNodesChannel = "fullnodes-channel"
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

	StartNode("4201", true, func(net *Network) {
		s.P2p = net
		//go jsonrpc.StartServer(cli, rpc, rpcPort, rpcAddr)
	})
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
	// providers := s.node.ProvidersFor("karai-xeq")
	// if providers != nil && len(providers) > 0 {
	// 	for _, provider := range providers {
	// 		if provider.GetID().Pub.String() == id.Pub.String() {
	// 			return provider
	// 		}
	// 	}
	// }
	return nil
}

/*

LookForNodes = Look for peers not known

*/
func (s *Server) LookForNodes() {
	// for {
	// 	if s.pl.Count < 9 {

	// 		providers := s.node.ProvidersFor("karai-xeq")
	// 		for _, provider := range providers {
	// 			go s.SendVersion(provider)
	// 		}
	// 	}

	// 	time.Sleep(10 * time.Second)
	// }
}

//NewDataTxFromCore = Go through all contracts and send data out
func (s *Server) NewDataTxFromCore(request []string, height int64, pubkey string) {

	var agg float64
	var tx transaction.Transaction
	var contract contract.Contract

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRowx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_hash=$1 ORDER BY tx_time DESC", request[0]).StructScan(&tx)

	err := json.Unmarshal([]byte(tx.Data), &contract)
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
		oracledata.Height = height
		oracledata.Pubkey = pubkey
		oracledata.Price = agg
		oracledata.Contract = tx.Hash
		oracledata.Hash = ""
		oracledata.Signature = ""
		sig, hash := api.CoreSign(oracledata)

		oracledata.Hash = hash
		oracledata.Signature = sig

		s.BroadCastOracleData(oracledata)
	}
}

//NewConsensusTXFromCore = create v1 tx
func (s *Server) NewConsensusTXFromCore(req transaction.NewBlock) {
	req_string, _ := json.Marshal(req)

	height := req.Height
	if height%10 != 0 {
		return
	}

	go s.Prtl.Mempool.PruneHeight(height - 5)

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	new_tx := transaction.CreateTransaction("1", txPrev, req_string, []string{}, []string{}, height)
	if !s.Prtl.Dat.HaveTx(new_tx.Hash) {
		s.Prtl.Dat.CommitDBTx(new_tx)
		s.BroadCastTX(new_tx)
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

	tx := transaction.CreateTransaction("3", txPrev, []byte(jsonContract), []string{}, []string{}, 0)

	if !s.Prtl.Dat.HaveTx(tx.Hash) {
		s.Prtl.Dat.CommitDBTx(tx)
		go s.BroadCastTX(tx)
	}
	log.Info("Created Contract " + tx.Hash[:8])
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
func (s *Server) CreateTrustedData(block_height int64) {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	contract_data_map := s.Prtl.Mempool.SortOracleDataMap(block_height)

	filtered_data_map, trusted_data_map := FilterOracleDataMap(contract_data_map)

	for _, contract_array := range filtered_data_map {
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
					log.Info("Unable to query last contract ref!")
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
			send := false
			if contract.Threshold != "" {
				log.Info(contract.Threshold)
				s, _ := strconv.ParseFloat(contract.Threshold, 64)
				if s > 0.0 {
					ltd := transaction.Trusted_Data{}
					json.Unmarshal([]byte(lastTrustedTx.Data), &ltd)

					change := PercentageChange(ltd.TrustedAnswer, price)

					if change >= s {
						send = true
					}
				}
			} else {
				send = true
			}

			if send {
				trusted_data := transaction.Trusted_Data{contract_array, price}

				new_tx := transaction.CreateTrustedTransaction(prev, trusted_data)

				s.Prtl.Dat.CommitDBTx(new_tx)
				s.BroadCastTX(new_tx)
			}
		}
	}
}

func StartNode(listenPort string, fullNode bool, callback func(*Network)) {
	// var r io.Reader
	// r = rand.Reader
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Creates a new RSA key pair for this host.
	prvKey, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)

	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(ws.New),
	)

	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

	// security := libp2p.Security(secio.ID, secio.New)
	if len(listenPort) == 0 {
		listenPort = "0"
	}

	listenAddrs := libp2p.ListenAddrStrings(
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s", listenPort),
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%s/ws", listenPort),
	)

	host, err := libp2p.New(
		ctx,
		transports,
		listenAddrs,
		muxers,
		libp2p.Identity(prvKey),
	)
	if err != nil {
		panic(err)
	}
	for _, addr := range host.Addrs() {
		fmt.Println("Listening on", addr)
	}
	log.Info("Host created: ", host.ID())

	// create a new PubSub service using the GossipSub router for general room
	pub, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}

	generalChannel, _ := JoinChannel(ctx, pub, host.ID(), GeneralChannel, true)
	subscribe := false
	subscribe = true
	miningChannel, _ := JoinChannel(ctx, pub, host.ID(), MiningChannel, subscribe)

	subscribe = false
	if fullNode {
		subscribe = true
	}
	fullNodesChannel, _ := JoinChannel(ctx, pub, host.ID(), FullNodesChannel, subscribe)

	// setup peer discovery
	err = SetupDiscovery(ctx, host)
	if err != nil {
		panic(err)
	}
	network := &Network{
		Host:             host,
		GeneralChannel:   generalChannel,
		MiningChannel:    miningChannel,
		FullNodesChannel: fullNodesChannel,
		Transactions:     make(chan *transaction.Transaction, 200),
		// Miner:            miner,
	}
	callback(network)
	//err = RequestBlocks(network)

	// go HandleEvents(network)
	// if miner {
	// 	// event loop for miners to constantly send a ping to fullnodes for new transactions
	// 	// in order for it to be mined and added to the blockchain
	// 	go network.MinersEventLoop()
	// }

	if err != nil {
		panic(err)
	}
	// if err = ui.Run(network); err != nil {
	// 	log.Error("error running text UI: %s", err)
	// }
}

// func HandleEvents(net *Network) {
// 	for {
// 		select {
// 		case block := <-net.Blocks:
// 			net.SendBlock("", block)
// 		case tnx := <-net.Transactions:
// 			// mine := false
// 			net.SendTx("", tnx)
// 		}
// 	}
// }
// func RequestBlocks(net *Network) error {
// 	peers := net.GeneralChannel.ListPeers()
// 	// Send version
// 	if len(peers) > 0 {
// 		net.SendVersion(peers[0].Pretty())
// 	}
// 	return nil
// }

func SetupDiscovery(ctx context.Context, host host.Host) error {

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	// Bootstrap the DHT. In the default configuration, this spawns a Background
	// thread that will refresh the peer table every five minutes.
	log.Info("Bootstrapping the DHT")
	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	// Let's connect to the bootstrap nodes first. They will tell us about the
	// other nodes in the network.

	var wg sync.WaitGroup
	for _, peerAddr := range dht.DefaultBootstrapPeers {
		peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerinfo); err != nil {
				log.Error(err)
			} else {
				log.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	// We use a rendezvous point "meet me here" to announce our location.
	// This is like telling your friends to meet you at the Eiffel Tower.
	log.Info("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, "rendezvous")
	log.Info("Successfully announced!")

	// Now, look for others who have announced
	// This is like your friend telling you the location to meet you.
	log.Info("Searching for other peers...")
	peerChan, err := routingDiscovery.FindPeers(ctx, "rendezvous")
	if err != nil {
		panic(err)
	}

	// Finally we open streams to the newly discovered peers.
	for peer := range peerChan {
		if peer.ID == host.ID() {
			continue
		}
		log.Debug("Found peer:", peer)

		log.Debug("Connecting to:", peer)
		err := host.Connect(context.Background(), peer)
		if err != nil {
			log.Warningf("Error connecting to peer %s: %s\n", peer.ID.Pretty(), err)
			continue
		}
		log.Info("Connected to:", peer)
	}

	return nil
}
