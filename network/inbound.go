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
	// "time"
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
		//nothing

	} else {


		db, connectErr := s.Prtl.Dat.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)


		transactions := []transaction.Transaction{}
		lhash := last_hash
		hit := false
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

			if lhash == this_tx.Hash {
				log.Println("hit")
				hit = true
			}

			if hit == true {
				transactions = append(transactions, this_tx)

				//loop through to find contracts
				row2, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='3' AND tx_prnt=$1 ORDER BY tx_time ASC", this_tx.Hash)
				if err != nil {
					panic(err)
				}
				defer row2.Close()
				for row2.Next() {
					var t_tx transaction.Transaction
					err = row2.StructScan(&t_tx)
					if err != nil {
						// handle this error
						log.Panic(err)
					}
					transactions = append(transactions, t_tx)
					//loop through to find oracle data
						row3, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='2' AND tx_epoc=$1 ORDER BY tx_time ASC", t_tx.Epoc)
						if err != nil {
							panic(err)
						}
						defer row3.Close()
						for row3.Next() {
							var t2_tx transaction.Transaction
							err = row3.StructScan(&t2_tx)
							if err != nil {
								// handle this error
								log.Panic(err)
							}
							transactions = append(transactions, t2_tx)
						}
						err = row3.Err()
						if err != nil {
							log.Panic(err)
						}
				}
				err = rows.Err()
				if err != nil {
					log.Panic(err)
				}				
			}
		}

		// get any error encountered during iteration
		err = rows.Err()
		if err != nil {
			log.Panic(err)
		}
		var txes [][]byte
		for _, tx := range transactions {
				txes = append(txes, tx.Serialize())
		}

		data := GOB_BATCH_TX{txes}
		payload := GobEncode(data)
		request := append(CmdToBytes("batchtx"), payload...)

		go s.SendData(ctx, request);
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

func (s *Server) HandleBatchTx(ctx *flatend.Context, request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GOB_BATCH_TX

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}
	
	for _, tx_ := range payload.Batch {
		
		tx := transaction.DeserializeTransaction(tx_)

		log.Println("[RECV] [" + command + "] Received " + strconv.Itoa(len(payload.Batch)) + " Transactions")


		if s.Prtl.Dat.HaveTx(tx.Prev) {
			if !s.Prtl.Dat.HaveTx(tx.Hash) {
				s.Prtl.Dat.CommitDBTx(tx)
				json_tx, _ := json.Marshal(tx)
				for _, conn := range s.Sockets {
					if err := conn.WriteMessage(1, json_tx); err != nil {
						log.Println(err)
						return
					}
				}
			}
		}
	}


	//s.SendInv("tx", [][]byte{[]byte(tx.Hash)})		
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


	if s.Prtl.Dat.HaveTx(tx.Prev) {
		if !s.Prtl.Dat.HaveTx(tx.Hash) {
			s.Prtl.Dat.CommitDBTx(tx)

			json_tx, _ := json.Marshal(tx)
			for _, conn := range s.Sockets {
				if err := conn.WriteMessage(1, json_tx); err != nil {
					log.Println(err)
					return
				}
			}
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
	case "batchtx":
		go s.HandleBatchTx(ctx, req)
	case "version":
		go s.HandleVersion(ctx, req)
	}

}