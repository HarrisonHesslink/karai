package network

import (	
	"bytes"
	"encoding/gob"
	"log"
	"strconv"
	//"io/ioutil"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"github.com/harrisonhesslink/flatend"
	"time"
	"encoding/json"
)
func (s *Server) HandleAddr(request []byte) {
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

func (s *Server) HandleGetTxes(ctx *flatend.Context, request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GetTxes

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	log.Println("[RECV] [" + command + "] Get Tx from: " + payload.Top_hash)

	top_hash, _  := s.Prtl.Dat.ReturnTopHash()
	if top_hash != "" {
		//somethin
	}

	if !s.Prtl.Dat.HaveTx(payload.Top_hash) {
		//somethin
	}

	var tx_hash string 
	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()

	util.Handle("Error creating a DB connection: ", connectErr)
	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_prev=$1 AND tx_type=$2",payload.Top_hash, "1").Scan(&tx_hash)

	next_tx := s.Prtl.Dat.GetTransaction([]byte(tx_hash))
	log.Println(tx_hash)
	s.SendTx(s.GetProviderFromID(&ctx.ID), next_tx)

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

	log.Println("[RECV] [" + command + "] Transaction: " + tx.Hash)

	if tx.Type == "2" {
		var txData string
		db, connectErr := s.Prtl.Dat.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)

		var this_tx_data transaction.Request_Data_TX
		err := json.Unmarshal([]byte(tx.Data), &this_tx_data)
		if err != nil {
			// handle this error
			log.Panic(err)
		}

		i := 0
		for i <= 10 {
			log.Println("[SELF] [" + command + "] Trying to add: " + tx.Hash)

			_ = db.QueryRow("SELECT tx_data FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txData)

			var last_consensus_data transaction.Request_Data_TX
			err := json.Unmarshal([]byte(txData), &txData)
			if err != nil {
				// handle this error
				log.Panic(err)
			}

			if last_consensus_data.Height == this_tx_data.Height {
				if !s.Prtl.Dat.HaveTx(tx.Hash) {
					s.Prtl.Dat.CommitDBTx(tx)
				}
				break;
			}

			i++
			time.Sleep(5 * time.Second)
		}

	} else {
		if !s.Prtl.Dat.HaveTx(tx.Hash) {
			s.Prtl.Dat.CommitDBTx(tx)
		}
	}

	//s.SendInv("tx", [][]byte{[]byte(tx.Hash)})		

}

func (s *Server) HandleVersion(ctx *flatend.Context, request []byte) {
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
		s.SendGetTxes(ctx)
	}

	log.Println("[RECV] [" + command + "] Node has Num Tx: " + strconv.Itoa(payload.TxSize))
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
		s.HandleVersion(ctx, req)
	default:
		log.Println("Unknown command")
	}

}