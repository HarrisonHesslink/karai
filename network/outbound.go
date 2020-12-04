package network

import (
	"bytes"

	"github.com/harrisonhesslink/flatend"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"

	//"math/rand"
	//"time"
	"io/ioutil"
	//"github.com/lithdew/kademlia"
)

// func(s *Server) RequestTxes(ctx *flatend.Context) {
// 	for _, node := range KnownNodes {
// 		/s.SendGetTxes(node)
// 	}
// }

// func (s Server) SendAddr(provider flatend.Provider) {
// 	nodes := Addr{KnownNodes}
// 	nodes.AddrList = append(nodes.AddrList, nodeAddress)
// 	payload := GobEncode(nodes)
// 	request := append(CmdToBytes("addr"), payload...)

// 	s.SendData(address, request)
// }

/*

SendTx : Send singular tx

*/
func (s *Server) SendTx(p *flatend.Provider, tx transaction.Transaction) {
	data := GOB_TX{tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	_, err := p.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
	if err == nil {
		util.Success_log(util.Send + " [TXT] Sending Transaction to " + p.GetID().Pub.String() + " ip: " + p.GetID().Host.String())
	}
}

/*

BroadCastTX : broadcast tx

*/
func (net *Network) BroadCastTX(tx transaction.Transaction) {
	data := GOB_TX{tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	net.GeneralChannel.Publish("Recieved NEW TX: ", request, "")
}

/*

BroadCastOracleData : broadcast oracle data

*/
func (net *Network) BroadCastOracleData(oracle_data transaction.OracleData) {
	data := GOB_ORACLE_DATA{oracle_data}
	payload := GobEncode(data)
	request := append(CmdToBytes("data"), payload...)

	net.GeneralChannel.Publish("Recieved NEW ORACLE DATA", request, "")
}

/*

SendData : sendBytes

*/
func (s *Server) SendData(ctx *flatend.Context, data []byte) {

	p := s.GetProviderFromID(&ctx.ID)
	if p != nil {
		// stream, err := s.node.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
		// if err == nil {
		// 	go s.HandleCall(stream)
		// }
	}
}

/*

BroadCastData : Broadcast data

*/
func (s *Server) BroadCastData(data []byte) {
	// providers := s.node.ProvidersFor("karai-xeq")
	// for _, provider := range providers {
	// 	_, err := provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
	// 	if err != nil {
	// 		//fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
	// 	}
	// }
}

/*

SendGetTxes : Get tansactions not known

*/
func (s *Server) SendGetTxes(ctx *flatend.Context, fill bool, contracts map[string]string) {

	var txPrev string

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)
	payload := GobEncode(GetTxes{txPrev, fill, contracts})
	request := append(CmdToBytes("gettxes"), payload...)

	go s.SendData(ctx, request)

	if !fill {
		util.Success_log(util.Send + " [GTXS] Requesting Transactions starting from: " + txPrev)
	} else {
		util.Success_log(util.Send + " [GTXS] Requesting Contracts and Data")
	}
}

/*

SendVersion : Send Sync Call

*/
func (n *Network) SendVersion() {

	// db, connectErr := s.Prtl.Dat.Connect()
	// defer db.Close()
	// util.Handle("Error creating a DB connection: ", connectErr)

	// var txPrev string
	// _ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	// contracts := s.GetContractMap()

	payload := GobEncode(SyncCall{"yemen", nil})

	request := append(CmdToBytes("version"), payload...)

	n.GeneralChannel.Publish("Recieved send version", request, "")
}
