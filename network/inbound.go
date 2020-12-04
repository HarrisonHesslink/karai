package network

import (
	"bytes"
	"encoding/gob"

	// "fmt"
	"log"
	"strconv"

	"github.com/harrisonhesslink/flatend"

	//"io/ioutil"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"

	// "time"
	"encoding/json"
)

/*
This function handles request for transactions. It takes a top consensus tx hash. 100 txes per batch
*/
func (s *Server) HandleGetTxes(ctx *flatend.Context, request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GetTxes

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Println("ERROR HandleGetTxs: Failed to decode payload", err)
		return
	}

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	transactions := []transaction.Transaction{}

	if payload.FillData {
		for key, value := range payload.Contracts {

			//loop through to find oracle data
			row3, err := db.Queryx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_type='2' AND tx_epoc=$1 ORDER BY tx_time DESC", key)
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

				if value == t2_tx.Hash {
					row3.Close()
					break
				}

				transactions = append(transactions, t2_tx)
			}
			err = row3.Err()
			if err != nil {
				log.Panic(err)
			}

			if value == "need" {
				var contract_tx transaction.Transaction
				_ = db.QueryRowx("SELECT * FROM "+s.Prtl.Dat.Cf.GetTableName()+" WHERE tx_hash=$1", key).StructScan(&contract_tx)
				transactions = append(transactions, contract_tx)
				continue
			}
		}

	} else {
		util.Success_log(util.Rcv + " [" + command + "] Get Tx from: " + payload.Top_hash)

		if s.Prtl.Dat.HaveTx(payload.Top_hash) {
			rows, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC")
			if err != nil {
				log.Println("Error query")
				return
			}
			for rows.Next() {
				var this_tx transaction.Transaction
				err = rows.StructScan(&this_tx)
				if err != nil {
					log.Println(err)
					log.Println("Error query")
					break
				}

				if this_tx.Hash == payload.Top_hash {
					rows.Close()
					break
				}

				transactions = append(transactions, this_tx)
			}
		}
	}

	var txes [][]byte

	for i := len(transactions) - 1; i >= 0; i-- {

		txes = append(txes, transactions[i].Serialize())
		if (i % 100) == 0 {
			data := GOB_BATCH_TX{txes, len(transactions)}
			payload := GobEncode(data)
			request := append(CmdToBytes("batchtx"), payload...)

			s.SendData(ctx, request)
			txes = [][]byte{}
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
		log.Println("Unable to decode")
		return
	}

	if payload.Type == "tx" {

		//tx := s.Prtl.Dat.GetTransaction(payload.ID)

		//s.SendTx(s.GetProviderFromID(&ctx.ID), tx)
	}
	util.Success_log(util.Rcv + " [" + command + "] Data Request from: " + ctx.ID.Pub.String())
}

func (s *Server) HandleBatchTx(ctx *flatend.Context, request []byte) {
	command := BytesToCmd(request[:commandLength])

	var buff bytes.Buffer
	var payload GOB_BATCH_TX

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Println("Unable to decode")
		return
	}

	db, connectErr := s.Prtl.Dat.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	count := 0

	if !s.isSyncing {
		s.isSyncing = true
		for _, tx_ := range payload.Batch {

			tx := transaction.DeserializeTransaction(tx_)
			if s.Prtl.Dat.HaveTx(tx.Prev) {
				if !s.Prtl.Dat.HaveTx(tx.Hash) {
					count++
					s.Prtl.Dat.CommitDBTx(tx)
				}
			}
		}
		s.isSyncing = false
	}

	// if need_fill {
	// 	go s.SendGetTxes(ctx, true, s.GetContractMap())
	// }

	// percentage_float := float64(payload.TotalSent) / float64(s.tx_need) * 100
	// percentage_string := fmt.Sprintf("%.2f", percentage_float)
	if count > 0 {
		util.Success_log(util.Rcv + " [" + command + "] Received Transactions: " + strconv.Itoa(count)) //. Sync %:" + percentage_string + "[" + strconv.Itoa(payload.TotalSent) + "/" + strconv.Itoa(s.tx_need) + "]")
	}
	// if payload.TotalSent == s.tx_need {
	// 	s.tx_need = 0
	// 	s.sync = false
	// }
}

func (net *Network) HandleTx(content *ChannelContent) {
	command := BytesToCmd(content.Payload[:commandLength])

	var buff bytes.Buffer
	var payload GOB_TX

	buff.Write(content.Payload[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Println("Unable to decode")
		return
	}
	txData := payload.TX
	tx := transaction.DeserializeTransaction(txData)

	if tx.Type == "1" {
		var consensus_data transaction.Request_Consensus
		err := json.Unmarshal([]byte(tx.Data), &consensus_data)
		if err != nil {
			return
		}

		if net.Database.HaveTx(tx.Prev) {
			if !net.Database.HaveTx(tx.Hash) {
				net.Database.CommitDBTx(tx)
				net.BroadCastTX(tx)
			}
		}

	} else {
		if net.Database.HaveTx(tx.Prev) {
			if !net.Database.HaveTx(tx.Hash) {
				net.Database.CommitDBTx(tx)

				oracleData := transaction.OracleData{}

				json.Unmarshal([]byte(tx.Data), &oracleData)

				net.BroadCastTX(tx)
			}
		}
	}
	net.Ui.displaySelfMessage(" [" + command + "] Transaction: " + tx.Hash)
}

// func (s *Server) HandleData(ctx *flatend.Context, request []byte) {
// 	command := BytesToCmd(request[:commandLength])

// 	var buff bytes.Buffer
// 	var payload GOB_ORACLE_DATA

// 	buff.Write(request[commandLength:])
// 	dec := gob.NewDecoder(&buff)
// 	err := dec.Decode(&payload)
// 	if err != nil {
// 		log.Println("Unable to decode")
// 		return
// 	}

// 	if s.Prtl.Mempool.addOracleData(payload.Oracle_Data) {
// 		s.BroadCastOracleData(payload.Oracle_Data)
// 		util.Success_log(util.Rcv + " [" + command + "] Oracle Data: " + payload.Oracle_Data.Hash)
// 	}
// }

// func (net *Network) HandleSyncCall(content *ChannelContent) {
// 	command := BytesToCmd(content.Payload[:commandLength])
// 	var buff bytes.Buffer
// 	var payload SyncCall

// 	buff.Write(request[commandLength:])
// 	dec := gob.NewDecoder(&buff)
// 	err := dec.Decode(&payload)
// 	if err != nil {
// 		log.Println("Unable to decode")
// 		return
// 	}

// 	db, connectErr := s.Prtl.Dat.Connect()
// 	defer db.Close()
// 	util.Handle("Error creating a DB connection: ", connectErr)

// 	var txPrev string
// 	_ = db.QueryRow("SELECT tx_hash FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC").Scan(&txPrev)

// 	if txPrev == "" {
// 		return
// 	}

// 	var request_contracts map[string]string
// 	if s.Prtl.Dat.HaveTx(payload.TopHash) {
// 		//okay, our v1 txes are all synced, lets check v2/v3
// 		our_contracts := s.GetContractMap()

// 		request_contracts = make(map[string]string)

// 		for contract_hash, top_hash := range payload.Contracts {
// 			if _, ok := our_contracts[contract_hash]; !ok {
// 				//does not have v3 tx so add it to request payload
// 				request_contracts[contract_hash] = "need"
// 				continue
// 			}

// 			if !containsValue(our_contracts, top_hash) {
// 				if !s.Prtl.Dat.HaveTx(top_hash) {
// 					//we shouldn't have this tx
// 					request_contracts[contract_hash] = our_contracts[contract_hash]
// 					continue
// 				}
// 			}

// 		}

// 		go s.SendGetTxes(ctx, true, request_contracts)
// 	} else {
// 		//get v1
// 		go s.SendGetTxes(ctx, false, request_contracts)
// 	}

// 	util.Success_log(util.Rcv + " [" + command + "]")
// }
