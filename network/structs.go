package network

import (
	"github.com/gorilla/websocket"
	"github.com/harrisonhesslink/flatend"
	config "github.com/harrisonhesslink/pythia/configuration"
	"github.com/harrisonhesslink/pythia/database"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/lithdew/kademlia"
)

type Addr struct {
	AddrList []string
}

type GOB_ORACLE_DATA struct {
	Oracle_Data transaction.OracleData
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

type NeedTX struct {
	Hash string
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

	Nodes []string

	isSyncing bool
	senate    *senate.Senate
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
	transactions []transaction.OracleData
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
