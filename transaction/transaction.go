package transaction

import (
	"bytes"
	"encoding/gob"
	"log"
	"github.com/karai/go-karai/util"
	"golang.org/x/crypto/sha3"
	"encoding/hex"
	"encoding/json"
)

// Transaction This is the structure of the transaction
type Transaction struct {
	Time string `json:"time" db:"tx_time"`
	Type string `json:"type" db:"tx_type"`
	Hash string `json:"hash" db:"tx_hash"`
	Data string `json:"data" db:"tx_data"`
	Prev string `json:"prev" db:"tx_prev"`
	Epoc string `json:"epoc" db:"tx_epoc"`
	Subg string `json:"subg" db:"tx_subg"`
	Prnt string `json:"prnt" db:"tx_prnt"`
	Mile bool   `json:"mile" db:"tx_mile"`
	Lead bool   `json:"lead" db:"tx_lead"`
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}
	return transaction
}

func CreateTransaction(txType, prev string, data string) Transaction {
	var newTx Transaction

	newTx.Type = txType
	// if isCoordinator && txType == "2" {
	if newTx.Type == "2" {
		parsePayload := json.Valid([]byte(data))
		if !parsePayload {
			newTx.Data = hex.EncodeToString([]byte(data))
		} else if parsePayload {
			newTx.Data = data
		}

		// db, connectErr := d.Connect()
		// defer db.Close()
		// util.Handle("Error creating a DB connection: ", connectErr)

		// _ = db.QueryRow("SELECT tx_hash FROM " + d.c.GetTableName() + " ORDER BY tx_time DESC LIMIT 1").Scan(&txPrev)

		newTx.Time = util.UnixTimeStampNano()
		newTx.Hash = hashTransaction(newTx.Time, newTx.Type, newTx.Data, prev)
		newTx.Epoc = "0"
		newTx.Mile = false

		// if d.txCount == 0 {
		// 	txLead = true
		// 	txSubg = txHash
		// 	txPrnt = txEpoc
		// 	d.thisSubgraph = txHash
		// 	txPrnt = d.thisSubgraph
		// 	d.thisSubgraphShortName = d.thisSubgraph[0:4]
		// 	go d.newSubGraphTimer()
		// } else if d.txCount > 0 {
		// 	txLead = false
		// 	txPrnt = d.thisSubgraph
		// 	txSubg = d.thisSubgraph
		// 	d.thisSubgraphShortName = d.thisSubgraph[0:4]
		// }
		log.Println("[SELF] New Transaction: " + newTx.Hash)
		return newTx
	}
	return newTx
}

// hashTransaction takes elements of a transaction and computes a hash using SHA512
func hashTransaction(txTime, txType, txData, txPrev string) string {
	hashedData := []byte(txTime + txType + txData + txPrev)
	slot := make([]byte, 64)
	sha3.ShakeSum256(slot, hashedData)
	// fmt.Printf("%x\n", slot)
	txHash := hex.EncodeToString(slot[:])
	// legacy sha512
	// hash := sha512.Sum512(hashedData)
	// txHash := hex.EncodeToString(hash[:])

	return txHash
}
