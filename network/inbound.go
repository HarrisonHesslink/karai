package network

import (	
	"bytes"
	"encoding/gob"
	"log"
	"strconv"
	//"io/ioutil"
	"github.com/karai/go-karai/transaction"
	"github.com/harrisonhesslink/flatend"
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
			have_tx := s.Prtl.Dat.HaveTx(string(byte_data))
			if !have_tx && len(byte_data) > 0 {
				//s.SendGetData(payload.AddrFrom, "tx", byte_data)
			}
		}

		log.Println("[RECV] [" + command + "] [" + payload.Type + "] Inventory call: " + payload.AddrFrom + " Inventory Items: " + strconv.Itoa(len(payload.Items)))	
	}
}

func (s Server) HandleGetTxes(ctx *flatend.Context, request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GetTxes

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	txSize := s.Prtl.Dat.GetDAGSize()

	if txSize <= payload.numTx {

		//txes := s.Prtl.Dat.ReturnRangeOfTransactions(payload.numTx)

		//s.SendInv(payload.AddrFrom, "tx", txes)
	} else {
		//ahead or equal nothing to do maybe relay? 
	}

	log.Println("[RECV] [" + command + "] Get Txes from: " + ctx.ID.Pub.String())
}

func (s *Server) HandleGetData(ctx *flatend.Context, request []byte) {
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

		//tx := s.Prtl.Dat.GetTransaction(payload.ID)

		//s.SendTx(s.GetProviderFromID(&ctx.ID), tx)
	}
	log.Println("[RECV] [" + command + "] Data Request from: " + ctx.ID.Pub.String())
}

func (s *Server) HandleTx(ctx *flatend.Context, request []byte) {
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

	s.SendInv("tx", [][]byte{[]byte(tx.Hash)})		

	if !s.Prtl.Dat.HaveTx(tx.Hash) {
		s.Prtl.Dat.CommitDBTx(tx)
	}

	log.Println("[RECV] [" + command + "] Handle Transaction: " + tx.Hash)
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

	if payload.TxSize > s.Prtl.Dat.GetDAGSize() {
		//s.RequestTxes()
	}
	log.Println("[RECV] [" + command + "] Num Tx: " + strconv.Itoa(payload.TxSize))
}

func (s *Server) HandleConnection(req []byte, ctx *flatend.Context) {

	command := BytesToCmd(req[:commandLength])
	switch command {
	case "addr":
		s.HandleAddr(req)
	case "inv":
		s.HandleInv(req)
	 case "gettxes":
	 	s.HandleGetTxes(ctx, req)
	 case "getdata":
	 	s.HandleGetData(ctx, req)
	case "tx":
		s.HandleTx(ctx, req)
	case "version":
		s.HandleVersion(req)
	default:
		log.Println("Unknown command")
	}

}