package network

import (
	"bytes"

	"github.com/harrisonhesslink/flatend"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
	log "github.com/sirupsen/logrus"

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
func (net *Network) SendData(data []byte) {
	net.GeneralChannel.Publish("Recieved NEW ORACLE DATA", data, "")
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
func (net *Network) SendGetTxes(fill bool, contracts map[string]string) {

	var txPrev string

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	_ = db.QueryRow("SELECT tx_hash FROM " + net.Database.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)
	payload := GobEncode(GetTxes{txPrev, fill, contracts})
	request := append(CmdToBytes("gettxes"), payload...)

	go net.SendData(request)

	if !fill {
		log.Info(util.Send + " [GTXS] Requesting Transactions starting from: " + txPrev)
	} else {
		log.Info(util.Send + " [GTXS] Requesting Contracts and Data")
	}
}

/*

SendVersion : Send Sync Call

*/
func (net *Network) SendVersion() {

	db, connectErr := net.Database.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var txPrev string
	_ = db.QueryRow("SELECT tx_hash FROM " + net.Database.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	contracts := net.GetContractMap()

	payload := GobEncode(SyncCall{txPrev, contracts})

	request := append(CmdToBytes("version"), payload...)

	net.GeneralChannel.Publish("Recieved send version", request, "")
}
