package peer_manager

import (
	config "github.com/karai/go-karai/configuration"
)

type Peer struct {

	PeerPubKey string
	PeerExternalIP string
	PeerExternalPort string

	FirstConnectTime int
	BytesSent int
	BytesReceived int 
	BanScore int
}

type PeerManager struct {
	Peers []Peer
	Cf *config.Config
	Count int
}