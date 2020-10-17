package network
import (
	"os"
	"github.com/karai/go-karai/db"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/peer_manager"
	"github.com/lithdew/flatend"

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
	AddrFrom string
	TX   []byte
}

type GetTxes struct {
	AddrFrom string
	numTx int
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
	prtl *Protocol
	cf *config.Config
	PeerManager *peer_manager.PeerManager
	node *flatend.Node

	Peers []string
	ExternalIP string
	ExternalPort int

	signalChannel chan os.Signal
}

type Protocol struct {
	dat *db.Database
}
