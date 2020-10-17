package network

import (
	//"github.com/karai/go-karai/transaction"
	"github.com/lithdew/flatend"
	//"bytes"
	"log"
	//"math/rand"
	//"time"
	//"strconv"
)

func(s Server) RequestTxes() {
	//for _, node := range KnownNodes {
		///s.SendGetTxes(node)
	//}
}

// func (s Server) SendAddr(provider flatend.Provider) {
// 	nodes := Addr{KnownNodes}
// 	nodes.AddrList = append(nodes.AddrList, nodeAddress)
// 	payload := GobEncode(nodes)
// 	request := append(CmdToBytes("addr"), payload...)

// 	s.SendData(address, request)
// }

// func(s Server)  SendTx(provider flatend.Provider, tx transaction.Transaction) {
// 	data := GOB_TX{tx.Serialize()}
// 	payload := GobEncode(data)
// 	request := append(CmdToBytes("tx"), payload...)

// //	s.SendData(addr, request)
// }

func (s Server) SendData(ctx *flatend.Context, data []byte) {
		ctx.Write(data)
}

// func (s Server) SendBroadcastTX(tx transaction.Transaction) {
// 	data := GOB_TX{tx.Serialize()}
// 	payload := GobEncode(data)
// 	request := append(CmdToBytes("broadtx"), payload...)

// 	rand.Seed(time.Now().UnixNano())
// 	// providers := s.node.ProvidersFor("karai-xeq")
// 	// for _, provider := range providers {
// 	// 	s.SendData(*provider, request)
// 	// 	// if err != nil {
// 	// 	// 	fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
// 	// 	// }
// 	// 	log.Println("[SEND] [BRD] Broadcasting Transaction: " + tx.Hash)
// 	// }
// }

// func (s Server) SendBroadcastNewPeer(provider flatend.Provider) {
// 	data := NewPeer{nodeAddress, addr}
// 	payload := GobEncode(data)
// 	request := append(CmdToBytes("newpeer"), payload...)

// 	rand.Seed(time.Now().UnixNano())

// 	ok := false
// 	//loop to make sure it broadcasts not self
// 	for ok == false {
// 		index := rand.Intn(len(KnownNodes))
// 		if KnownNodes[index] != nodeAddress {
// 			//s.SendData(KnownNodes[index], request)
// 			log.Println("[SEND] [BRD] Broadcasting Connected Peer: " + addr)
// 			ok = true
// 		}
// 	}
// }

// func (s Server) SendInv(provider flatend.Provider, kind string, items [][]byte) {
// 	inventory := Inv{nodeAddress, kind, items}
// 	payload := GobEncode(inventory)
// 	request := append(CmdToBytes("inv"), payload...)

// 	//s.SendData(address, request)
// 	log.Println("[SEND] [INV] Sending INV: " + strconv.Itoa(len(items)))
// }

// func (s Server)SendGetTxes(provider flatend.Provider) {
// 	payload := GobEncode(GetTxes{nodeAddress, s.prtl.dat.GetDAGSize()})
// 	request := append(CmdToBytes("gettxes"), payload...)

// 	//s.SendData(address, request)
// 	//log.Println("[SEND] [GTXS] Requesting Transactions to: " + address)
// }

// func (s Server) SendGetData(provider flatend.Provider, kind string, id []byte) {
// 	payload := GobEncode(GetData{nodeAddress, kind, id})
// 	request := append(CmdToBytes("getdata"), payload...)

// 	//s.SendData(address, request)
// 	//log.Println("[SEND] [GDTA][" + kind + "] Sending Data to: " + address)
// }

func (s Server) SendVersion(ctx *flatend.Context) {
	numTx := s.prtl.dat.GetDAGSize()

	payload := GobEncode(Version{version, numTx, s.ExternalIP})

	request := append(CmdToBytes("version"), payload...)

	s.SendData(ctx, request)
	log.Println("[SEND] [VERSION] Version Call")
}