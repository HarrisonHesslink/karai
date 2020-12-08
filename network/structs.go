package network

import (
	config "github.com/harrisonhesslink/pythia/configuration"
	"github.com/harrisonhesslink/pythia/database"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/libp2p/go-libp2p-core/host"
)

const (
	commandLength = 12
)

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

type SyncCall struct {
	TopHash   string
	Contracts map[string]string
}

type Server struct {
	cf    *config.Config
	Nodes []string
	P2p   *Network
}

type MemPool struct {
	transactions []transaction.OracleData
	// tx_hash -> block height
	transactions_map map[string]int
}

type ArrayTX struct {
	Txes []transaction.Transaction `json:txes`
}

type Network struct {
	Host             host.Host
	GeneralChannel   *Channel
	MiningChannel    *Channel
	FullNodesChannel *Channel
	Transactions     chan *transaction.Transaction
	Database         *database.Database
	Mempool          *MemPool
}
