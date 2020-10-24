package transaction

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"github.com/karai/go-karai/util"
	"golang.org/x/crypto/sha3"
	"log"
)

func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) ParseInterface() interface{} {
	if tx.Type == "1" {
		var tx_data Request_Consensus

		err := json.Unmarshal([]byte(tx.Data),&tx_data)
		if err != nil {
			log.Println("Unable to parse tx data")
			return nil
		}

		return tx_data
	} else if tx.Type == "2" {
		var tx_data Request_Oracle_Data
		err := json.Unmarshal([]byte(tx.Data),&tx_data)
		if err != nil {
			log.Println("Unable to parse tx data")
			return nil
		}

		return tx_data
	} else if tx.Type == "3" {
		var tx_data Request_Contract
		err := json.Unmarshal([]byte(tx.Data),&tx_data)
		if err != nil {
			log.Println("Unable to parse tx data")
			return nil
		}

		return tx_data
	}
	return nil
}

func CheckConsensusTx(consensus *Request_Consensus) bool {
	// isFound := false
	// for _, key := range last_consensus.Data {
	// 	if key == v.PubKey {
	// 		isFound = true
	// 		break
	// 	}
	// }

	// if !isFound {
	// 	return false
	// }

		return false


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

func CreateTransaction(txType, last_epoc_tx string, data []byte, txhash_on_epoc []string, txdata_on_epoc []string) Transaction {
	var newTx Transaction

	newTx.Type = txType
	// if isCoordinator && txType == "2" {
	if newTx.Type == "2" {
		parsePayload := json.Valid(data)
		if !parsePayload {
			newTx.Data = hex.EncodeToString(data)
		} else if parsePayload {
			newTx.Data = string(data)
		}

		rct := Request_Oracle_Data{}
		_ = json.Unmarshal(data, &rct)

		if last_epoc_tx == "" {
			newTx.Prev = rct.Epoc

		} else {
			newTx.Prev = last_epoc_tx
		}
		newTx.Time = util.UnixTimeStampNano()
		newTx.Epoc = rct.Epoc
		newTx.Mile = false

		newTx.Prnt = newTx.Epoc

		newTx.Hash = hashTransaction(newTx.Time, newTx.Type, newTx.Data, newTx.Prev)
		newTx.Subg = newTx.Epoc

		return newTx
	} else if newTx.Type == "1" {

		parsePayload := json.Valid(data)
		if !parsePayload {
			newTx.Data = hex.EncodeToString(data)
		} else if parsePayload {
			newTx.Data = string(data)
		}

		newTx.Prev = last_epoc_tx

		newTx.Time = util.UnixTimeStampNano()
		newTx.Hash = hashTransaction(newTx.Time, newTx.Type, newTx.Data, last_epoc_tx)
		newTx.Subg = newTx.Hash
		newTx.Epoc = "0"
		newTx.Mile = true
		newTx.Lead = true
		log.Println("[SELF] New Transaction: " + newTx.Hash)
		return newTx
	} else if newTx.Type == "3" {
		parsePayload := json.Valid(data)
		if !parsePayload {
			newTx.Data = hex.EncodeToString(data)
		} else if parsePayload {
			newTx.Data = string(data)
		}

		newTx.Prev = last_epoc_tx
		newTx.Time = util.UnixTimeStampNano()
		newTx.Hash = hashTransaction(newTx.Time, newTx.Type, newTx.Data, last_epoc_tx)
		newTx.Subg = newTx.Hash
		newTx.Epoc = newTx.Hash
		newTx.Mile = true
		newTx.Lead = false
		newTx.Prnt = last_epoc_tx
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
