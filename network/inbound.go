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
	last_hash := payload.Top_hash


	if !s.Prtl.Dat.HaveTx(last_hash) {

		//uh oh we have no hashy wot do?
		//lets request... 
		s.SendGetTxes(ctx)
		
	} else {


		db, connectErr := s.Prtl.Dat.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)


		transactions := []transaction.Transaction{}
		//Grab all first txes on epoc 
		rows, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time ASC")
		if err != nil {
			panic(err)
		}
		defer rows.Close()
		for rows.Next() {
			var this_tx transaction.Transaction
			err = rows.StructScan(&this_tx)
			if err != nil {
				// handle this error
				log.Panic(err)
			}
			transactions = append(transactions, this_tx)
		}
		// get any error encountered during iteration
		err = rows.Err()
		if err != nil {
			log.Panic(err)
		}

		for _, tx := range transactions {
			log.Println("Last Hash: " + last_hash)
			log.Println(tx.Prev)
			if tx.Prev == last_hash {
				p := s.GetProviderFromID(&ctx.ID)

				go s.SendTx(p, tx);
				last_hash = tx.Hash
			}	
		}
	}
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
	log.Println(request)
	txData := payload.TX
	tx := transaction.DeserializeTransaction(txData)

	log.Println("[RECV] [" + command + "] Transaction: " + tx.Hash)

	if tx.Type == "2" {
		db, connectErr := s.Prtl.Dat.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)

		this_tx_data := transaction.Request_Data_TX{}
		err := json.Unmarshal([]byte(tx.Data), &this_tx_data)
		if err != nil {
			// handle this error
			log.Panic(err)
		}

		i := 0
		for i <= 10 {
			var last_consensus_tx string
			var last_consensus_hash string
			log.Println("[SELF] [" + command + "] Trying to add: " + tx.Hash)

			_ = db.QueryRow("SELECT tx_hash, tx_data FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&last_consensus_hash, &last_consensus_tx)

			last_consensus_data := transaction.Request_Consensus_TX{}
			err := json.Unmarshal([]byte(last_consensus_tx), &last_consensus_data)
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
		go s.SendGetTxes(ctx)
	}

	log.Println("[RECV] [" + command + "] Node has Num Tx: " + strconv.Itoa(payload.TxSize))
}

func (s *Server) HandleConnection(req []byte, ctx *flatend.Context) {

	command := BytesToCmd(req[:commandLength])
	switch command {
	case "addr":
		go s.HandleAddr(req)
	case "inv":
		go s.HandleInv(req)
	 case "gettxes":
	 	go s.HandleGetTxes(ctx, req)
	 case "getdata":
	 	go s.HandleGetData(ctx, req)
	case "tx":
		go s.HandleTx(ctx, req)
	case "version":
		go s.HandleVersion(ctx, req)
	}

}