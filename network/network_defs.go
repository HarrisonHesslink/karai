package network
import (
	"os"
	"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/harrisonhesslink/flatend"
	"github.com/lithdew/kademlia"

)
const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	mineAddress     string
	KnownNodes      = []string{"127.0.0.1:3001"}
	txesInTransit = [][]byte{}
	txSize int
)

type Addr struct {
	AddrList []string
}

type GOB_TX struct {
	TX   []byte
}

type GOB_BATCH_TX struct {
	Batch [][]byte
}

type GetTxes struct {
	Top_hash string
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

type Version struct {
	Version    int
	TxSize int
	AddrFrom   string
}

type NewPeer struct {
	AddrFrom string
	NewPeer string
}

type Server struct { 
	Prtl *Protocol
	cf *config.Config
	node *flatend.Node
	pl *PeerList
	ExternalIP string
	ExternalPort int

	signalChannel chan os.Signal
}

type Protocol struct {
	Dat *db.Database
}

type Peer struct {
	ID *kademlia.ID
	Provider *flatend.Provider
}

type PeerList struct {
	Peers []Peer

	Count int 
}