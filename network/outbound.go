package network

import (
	"github.com/karai/go-karai/transaction"
	//"github.com/lithdew/kademlia"
	"fmt"
	"io/ioutil"
	"bytes"
	"log"
	"math/rand"
	"time"
	"strconv"
)

func(s Server) RequestTxes() {
	for _, node := range KnownNodes {
		s.SendGetTxes(node)
	}
}

func (s Server) SendAddr(address string) {
	nodes := Addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(CmdToBytes("addr"), payload...)

	s.SendData(address, request)
}

func(s Server)  SendTx(addr string, tx transaction.Transaction) {
	data := GOB_TX{nodeAddress, tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	if addr == "" {
		addr = nodeAddress
	}

	s.SendData(addr, request)
}

func (s Server) SendData(addr string, data []byte) {
	providers := s.node.ProvidersFor("karai-xeq")
	for _, provider := range providers {
		_, err := provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
		if err != nil {
			fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
		}
	}
}

func (s Server) SendBroadcastTX(tx transaction.Transaction) {
	data := GOB_TX{nodeAddress, tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("broadtx"), payload...)

	rand.Seed(time.Now().UnixNano())

	ok := false
	//loop to make sure it broadcasts not self
	for ok == false {
		index := rand.Intn(len(KnownNodes))
		if KnownNodes[index] != nodeAddress {
			s.SendData(KnownNodes[index], request)
			log.Println("[SEND] [BRD] Broadcasting Transaction: " + tx.Hash)
			ok = true
		}
	}
}

func (s Server) SendBroadcastNewPeer(addr string) {
	data := NewPeer{nodeAddress, addr}
	payload := GobEncode(data)
	request := append(CmdToBytes("newpeer"), payload...)

	rand.Seed(time.Now().UnixNano())

	ok := false
	//loop to make sure it broadcasts not self
	for ok == false {
		index := rand.Intn(len(KnownNodes))
		if KnownNodes[index] != nodeAddress {
			s.SendData(KnownNodes[index], request)
			log.Println("[SEND] [BRD] Broadcasting Connected Peer: " + addr)
			ok = true
		}
	}
}

func (s Server) SendInv(address, kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)

	s.SendData(address, request)
	log.Println("[SEND] [INV] Sending INV: " + strconv.Itoa(len(items)))
}

func (s Server)SendGetTxes(address string) {
	payload := GobEncode(GetTxes{nodeAddress, s.prtl.dat.GetDAGSize()})
	request := append(CmdToBytes("gettxes"), payload...)

	s.SendData(address, request)
	log.Println("[SEND] [GTXS] Requesting Transactions to: " + address)
}

func (s Server) SendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CmdToBytes("getdata"), payload...)

	s.SendData(address, request)
	log.Println("[SEND] [GDTA][" + kind + "] Sending Data to: " + address)
}

func (s Server) SendVersion(addr string) {
	numTx := s.prtl.dat.GetDAGSize()

	payload := GobEncode(Version{version, numTx, nodeAddress})

	request := append(CmdToBytes("version"), payload...)

	s.SendData(addr, request)
	log.Println("[SEND] [VERSION] Version Call")
}