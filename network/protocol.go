package network

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/harrisonhesslink/pythia/api"
	config "github.com/harrisonhesslink/pythia/configuration"
	contract "github.com/harrisonhesslink/pythia/contract"
	"github.com/harrisonhesslink/pythia/database"

	"encoding/json"
	"io/ioutil"
	"strconv"

	discovery "github.com/libp2p/go-libp2p-discovery"

	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
	_ "github.com/lib/pq"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	yamux "github.com/libp2p/go-libp2p-yamux"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
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
	s.RestAPI()

	StartNode("4201", true, func(net *Network) {
		s.P2p = net
		s.P2p.Database = database.NewDataBase(c)
		//go jsonrpc.StartServer(cli, rpc, rpcPort, rpcAddr)
	})
}

//NewDataTxFromCore = Go through all contracts and send data out
func (net *Network) NewDataTxFromCore(request []string, height int64, pubkey string) {

	var agg float64
	var tx transaction.Transaction
	var contract contract.Contract

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRowx("SELECT * FROM "+net.Database.Cf.GetTableName()+" WHERE tx_hash=$1 ORDER BY tx_time DESC", request[0]).StructScan(&tx)

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

		net.BroadCastOracleData(oracledata)
	}
}

//NewConsensusTXFromCore = create v1 tx
func (net *Network) NewConsensusTXFromCore(req transaction.NewBlock) {
	req_string, _ := json.Marshal(req)

	height := req.Height
	if height%10 != 0 {
		return
	}

	// go s.Prtl.Mempool.PruneHeight(height - 5)

	var txPrev string

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + net.Database.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	new_tx := transaction.CreateTransaction("1", txPrev, req_string, []string{}, []string{}, height)
	if !net.Database.HaveTx(new_tx.Hash) {
		net.Database.CommitDBTx(new_tx)
		net.BroadCastTX(new_tx)
	}
}

//CreateContract make new contract uploaded fron config.json
func (net *Network) CreateContract() {
	var txPrev string
	file, _ := ioutil.ReadFile("contract.json")

	data := contract.Contract{}

	_ = json.Unmarshal([]byte(file), &data)

	jsonContract, _ := json.Marshal(data)

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + net.Database.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	tx := transaction.CreateTransaction("3", txPrev, []byte(jsonContract), []string{}, []string{}, 0)

	if !net.Database.HaveTx(tx.Hash) {
		net.Database.CommitDBTx(tx)
		go net.BroadCastTX(tx)
	}

	message := "Contract Creation [v0.1.0 testnet]\n" +
		data.Asset + "/" + data.Denom + "\n" +
		"Explorer: https://pythia.equilibria.network"

	sendDiscordMessage("775986994551324694", message)
	sendTweet(message)

	log.Info("Created Contract " + tx.Hash[:8])
}

/*

GetContractMap creates contract map and their last known tx
*/
func (net *Network) GetContractMap() map[string]string {

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var Contracts map[string]string
	Contracts = make(map[string]string)

	//loop through to find oracle data
	rows, err := db.Queryx("SELECT * FROM " + net.Database.Cf.GetTableName() + " WHERE tx_type='3' ORDER BY tx_time DESC")
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
		_ = db.QueryRow("SELECT tx_hash FROM "+net.Database.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", this_tx.Hash).Scan(&data_prev)
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
func (net *Network) CreateTrustedData(block_height int64) {

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	contract_data_map := net.Mempool.SortOracleDataMap(block_height)

	filtered_data_map, trusted_data_map := FilterOracleDataMap(contract_data_map)

	for _, contract_array := range filtered_data_map {
		if len(contract_array) > 1 {

			var lastTrustedTx transaction.Transaction
			_ = db.QueryRowx("SELECT * FROM "+net.Database.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", contract_array[0].Contract).StructScan(&lastTrustedTx)
			prev := lastTrustedTx.Hash
			if prev == "" {
				return
			}

			multi := 1.0

			var contractTx transaction.Transaction
			_ = db.QueryRowx("SELECT * FROM "+net.Database.Cf.GetTableName()+" WHERE tx_hash=$1 ORDER BY tx_time DESC", contract_array[0].Contract).StructScan(&contractTx)

			contract := contract.Contract{}
			json.Unmarshal([]byte(contractTx.Data), &contract)

			if contract.ContractRef != "" {
				var lastContractRef transaction.Transaction
				_ = db.QueryRowx("SELECT * FROM "+net.Database.Cf.GetTableName()+" WHERE tx_epoc=$1 ORDER BY tx_time DESC", contract.ContractRef).StructScan(&lastContractRef)

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

				net.Database.CommitDBTx(new_tx)
				net.BroadCastTX(new_tx)

				// sf := fmt.Sprintf("%.12f", math.Round(trusted_data.TrustedAnswer*1000000000000)/1000000000000)

				// message := "Pythia Contract Request [v0.1.0 testnet]\n" +
				// 	sf + " " + contract.Asset + "/" + contract.Denom + "\n" +
				// 	strconv.Itoa(len(trusted_data.TrustedData)) + " Node Responses\nExplorer: https://pythia.equilibria.network"

				// sendDiscordMessage("775986994551324694", message)
				//sendTweet(message)
			}
		}
	}
}

func StartNode(listenPort string, fullNode bool, callback func(*Network)) {
	// var r io.Reader
	// r = rand.Reader
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prvkey, err := loadPeerKey()
	if err != nil {
		log.Error(err)
	}
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(ws.New),
	)

	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

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
		libp2p.EnableNATService(),
		libp2p.NATPortMap(),
		libp2p.Identity(prvkey),
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			400,         // HighWater,
			time.Minute, // GracePeriod
		)),
	)
	if err != nil {
		panic(err)
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
	go SetupDiscovery(ctx, host)

	if err != nil {
		panic(err)
	}

	network := &Network{
		Host:             host,
		GeneralChannel:   generalChannel,
		MiningChannel:    miningChannel,
		FullNodesChannel: fullNodesChannel,
		Transactions:     make(chan *transaction.Transaction, 200),
		Mempool:          NewMemPool(),
		// Miner:            miner,
	}
	callback(network)
	go network.hearbeat()

	if err != nil {
		panic(err)
	}
	network.handleEvents()
}

func (net *Network) hearbeat() {
	for {
		peers := net.GeneralChannel.ListPeers()
		for range peers {
			net.SendVersion()
		}
		time.Sleep(10 * time.Second)
	}
}

func RequestBlocks(net *Network) error {
	peers := net.GeneralChannel.ListPeers()
	for range peers {
		net.SendVersion()
	}
	return nil
}

func SetupDiscovery(ctx context.Context, host host.Host) {
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
				log.Debug("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	log.Debug("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(kademliaDHT)
	discovery.Advertise(ctx, routingDiscovery, "xeq-equilibria")
	log.Debug("P2P Network Initialized...")

	for {
		// Now, look for others who have announced
		// This is like your friend telling you the location to meet you.
		log.Debug("Searching for other peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, "xeq-equilibria")
		if err != nil {
			panic(err)
		}

		// Finally we open streams to the newly discovered peers.
		for peer := range peerChan {
			if peer.ID == host.ID() {
				continue
			}

			host.Peerstore().ClearAddrs(peer.ID)

			err := host.Connect(context.Background(), peer)
			if err != nil {
				continue
			}
			log.Info("Connected to:", peer)
		}
		time.Sleep(1 * time.Minute)
	}
}

func loadPeerKey() (crypto.PrivKey, error) {
	var prvkey crypto.PrivKey

	if _, err := os.Stat("peerkey"); err == nil {
		dat, _ := ioutil.ReadFile("peerkey")
		prvkey, err = crypto.UnmarshalEd25519PrivateKey(dat)
	} else if os.IsNotExist(err) {

		key, _, err := crypto.GenerateKeyPair(
			crypto.Ed25519,
			-1,
		)
		f, err := os.Create("peerkey")
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		key_bytes, _ := key.Raw()
		l, err := f.Write(key_bytes)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return nil, err
		}
		fmt.Println(l, "bytes written successfully")
		err = f.Close()
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
	}
	return prvkey, nil
}

func (net *Network) HandleStream(content *ChannelContent) {
	// ui.displayContent(content)
	if content.Payload != nil {
		command := BytesToCmd(content.Payload[:commandLength])

		switch command {
		case "gettxes":
			net.HandleGetTxes(content)
		case "tx":
			net.HandleTx(content)
		case "data":
			net.HandleData(content)
		case "batchtx":
			net.HandleBatchTx(content)
		case "version":
			net.HandleSyncCall(content)
		}
	}
}

func (net *Network) handleEvents() {
	var stopChan = make(chan os.Signal, 2)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	//go ui.readFromLogs(net.Blockchain.InstanceId)
	log.Info("HOST ADDR: ", net.Host.Addrs())

	for {
		select {
		case m := <-net.GeneralChannel.Content:
			net.HandleStream(m)
		case m := <-net.MiningChannel.Content:
			net.HandleStream(m)

		case m := <-net.FullNodesChannel.Content:
			net.HandleStream(m)
		case <-net.GeneralChannel.ctx.Done():
			return
		case <-stopChan:
			return
		}
	}
}
