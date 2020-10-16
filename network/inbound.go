package network

import (	
	"net"
	"bytes"
	"encoding/gob"
	"log"
	"strconv"
	"io/ioutil"
	"github.com/karai/go-karai/transaction"
)
func (s Server) HandleAddr(request []byte) {
	command := BytesToCmd(request[:commandLength])
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)

	}

	for _, n := range payload.AddrList {
		if !stringInSlice(n, KnownNodes) {
			KnownNodes = append(KnownNodes, n)

		}
	}

	log.Println("[RECV] [" + command + "] Known Nodes On Network: " + strconv.Itoa(len(KnownNodes)))	
}


func (s Server) HandleInv(request []byte) {
	command := BytesToCmd(request[:commandLength])
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "tx" {
		txesInTransit = payload.Items

		for _, byte_data := range payload.Items {
			have_tx := s.prtl.dat.HaveTx(string(byte_data))
			if !have_tx && len(byte_data) > 0 {
				s.SendGetData(payload.AddrFrom, "tx", byte_data)
			}
		}

		log.Println("[RECV] [" + command + "] [" + payload.Type + "] Inventory call: " + payload.AddrFrom + " Inventory Items: " + strconv.Itoa(len(payload.Items)))	
	}
}

func (s Server) HandleGetTxes(request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GetTxes

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	txSize := s.prtl.dat.GetDAGSize()

	if txSize <= payload.numTx {

		txes := s.prtl.dat.ReturnRangeOfTransactions(payload.numTx)

		s.SendInv(payload.AddrFrom, "tx", txes)
	} else {
		//ahead or equal nothing to do maybe relay? 
	}

	log.Println("[RECV] [" + command + "] Get Txes from: " + payload.AddrFrom)
}

func (s Server) HandleGetData(request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "tx" {

		tx := s.prtl.dat.GetTransaction(payload.ID)
		s.SendTx(payload.AddrFrom, tx)
	}
	log.Println("[RECV] [" + command + "] Data Request from: " + payload.AddrFrom)
}

func (s Server) HandleTx(request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GOB_TX

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.TX
	tx := transaction.DeserializeTransaction(txData)

	for _, node := range KnownNodes {
		if node != nodeAddress && node != payload.AddrFrom {
			s.SendInv(node, "tx", [][]byte{[]byte(tx.Hash)})
		}
	}
	log.Println("[RECV] [" + command + "] Handle Transaction: " + tx.Hash)

	if !s.prtl.dat.HaveTx(tx.Hash) {
		s.prtl.dat.CommitDBTx(tx)
	}
}

func (s Server) HandleVersion(request []byte) {
	command := BytesToCmd(request[:commandLength])
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	if !NodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
		s.SendBroadcastNewPeer(payload.AddrFrom)
	}

	if payload.TxSize > s.prtl.dat.GetDAGSize() {
		s.RequestTxes()
	}

	log.Println("[RECV] [" + command + "] Peers Known: " + strconv.Itoa(len(KnownNodes)) + " Num Tx: " + strconv.Itoa(payload.TxSize))
	s.SendAddr(payload.AddrFrom)
}

func (s Server) HandleToSync(request []byte) {

}

func (s Server) HandleNewPeer(request []byte) {
	command := BytesToCmd(request[:commandLength])
	var buff bytes.Buffer
	var payload NewPeer

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)

	}

	if !stringInSlice(payload.NewPeer, KnownNodes) {
		KnownNodes = append(KnownNodes, payload.NewPeer)
		s.SendVersion(nodeAddress)
	}
	
	log.Println("[RECV] [" + command + "] New Relayed Peer: " + payload.NewPeer)
}

func (s *Server) HandleConnection(conn net.Conn) {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()
	
	if err != nil {
		log.Panic(err)
	}
	command := BytesToCmd(req[:commandLength])

	switch command {
	case "addr":
		s.HandleAddr(req)
	case "inv":
		s.HandleInv(req)
	 case "gettxes":
	 	s.HandleGetTxes(req)
	 case "getdata":
	 	s.HandleGetData(req)
	case "tx":
		s.HandleTx(req)
	case "version":
		s.HandleVersion(req)
	case "broadtx":
		s.HandleTx(req)
	case "sync":
		s.HandleToSync(req)
	case "newpeer":
		s.HandleNewPeer(req)
	default:
		log.Println("Unknown command")
	}

}