package network

import (
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"github.com/harrisonhesslink/flatend"
	"bytes"
	//"math/rand"
	//"time"
	"strconv"
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

func(s *Server)  SendTx(p *flatend.Provider, tx transaction.Transaction) {
	data := GOB_TX{tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	_, err := p.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
	if err == nil {
		util.Success_log(util.Send + " [TXT] Sending Transaction to " + p.GetID().Pub.String() + " ip: " + p.GetID().Host.String())
	}
}

func(s *Server)  BroadCastTX(tx transaction.Transaction) {
	data := GOB_TX{tx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	_, err := s.node.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
	if err == nil {
		util.Success_log(util.Send + " [TXT] Broadcasting Transaction Out")
	}

}


func(s *Server)  BroadCastOracleData(oracle_data transaction.Request_Oracle_Data) {
	data := GOB_ORACLE_DATA{oracle_data}
	payload := GobEncode(data)
	request := append(CmdToBytes("data"), payload...)

	_, err := s.node.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
	if err == nil {
		util.Success_log(util.Send + " [DATA] Broadcasting Oracle Data Out")
	}

}

func (s *Server) SendData(ctx *flatend.Context, data []byte) {

	p := s.GetProviderFromID(&ctx.ID)
	if p != nil {
		stream, err := p.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
		if err == nil {
			go s.HandleCall(stream)
		}
	}
}

func (s *Server) BroadCastData(data []byte) {
	providers := s.node.ProvidersFor("karai-xeq")
	for _, provider := range providers {
		_, err := provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(data)))
		if err != nil {
			//fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
		}
	}
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

func (s *Server) SendInv( kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)

	for _, p := range s.pl.Peers {
			stream, err := p.Provider.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
			if err != nil {
				//fmt.Printf("Unable to broadcast to %s: %s\n", provider.Addr(), err)
			}
			util.Success_log(util.Send + " [INV] Sending INV: " + strconv.Itoa(len(items)))
			s.HandleCall(stream)
	}
}

func (s *Server)SendGetTxes(ctx *flatend.Context, fill bool, contracts map[string]string) {
	
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

// func (s Server) SendGetData(provider flatend.Provider, kind string, id []byte) {
// 	payload := GobEncode(GetData{nodeAddress, kind, id})
// 	request := append(CmdToBytes("getdata"), payload...)

// 	//s.SendData(address, request)
// 	//log.Println("[SEND] [GDTA][" + kind + "] Sending Data to: " + address)
// }

func (s *Server) SendVersion(p *flatend.Provider) {

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var txPrev string
	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

	contracts := s.GetContractMap()

	payload := GobEncode(SyncCall{txPrev, contracts})

	request := append(CmdToBytes("version"), payload...)

	stream, err := p.Push([]string{"karai-xeq"}, nil, ioutil.NopCloser(bytes.NewReader(request)))
	if err == nil {
		go s.HandleCall(stream)
		util.Success_log(util.Send + " [VERSION] Call")
	}

}