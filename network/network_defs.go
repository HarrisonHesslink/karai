package network

import (
	"github.com/gorilla/websocket"
	"github.com/harrisonhesslink/flatend"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/database"
	"github.com/karai/go-karai/transaction"
	"github.com/lithdew/kademlia"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress   string
	mineAddress   string
	KnownNodes    = []string{"127.0.0.1:3001"}
	txesInTransit = [][]byte{}
	txSize        int
)

type Addr struct {
	AddrList []string
}

type GOB_ORACLE_DATA struct {
	Oracle_Data transaction.Request_Oracle_Data
}

type GOB_TX struct {
	TX []byte
}

type GOB_BATCH_TX struct {
	Batch     [][]byte
	TotalSent int
}

type GetTxes struct {
	Top_hash string
	FillData bool

	//contract has keytype and top data hash as value
	Contracts map[string]string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type SyncCall struct {
	TopHash   string
	Contracts map[string]string
}

type NewPeer struct {
	AddrFrom string
	NewPeer  string
}

type Server struct {
	Prtl         *Protocol
	cf           *config.Config
	node         *flatend.Node
	pl           *PeerList
	ExternalIP   string
	ExternalPort int
	Sockets      []*websocket.Conn
}

type Protocol struct {
	Dat     *database.Database
	Sync    *Syncer
	Mempool *MemPool

	ConsensusNode     string
	LastConsensusNode string
	MyNodeKey         string
}

type MemPool struct {
	transactions []transaction.Request_Oracle_Data
	// tx_hash -> block height
	transactions_map map[string]int
}

type Syncer struct {
	//contracts = map[key = contract hash], value = top tx hash of data point
	Contracts map[string]string

	Last_tx   int
	Synced    bool
	Connected bool
	Tx_need   int
}

type Peer struct {
	ID       *kademlia.ID
	Provider *flatend.Provider
}

type PeerList struct {
	Peers []Peer

	Count int
}

type ArrayTX struct {
	Txes []transaction.Transaction `json:txes`
}

type ErrorJson struct {
	Message string `json:message`
	Error   bool   `json:is_error`
}
