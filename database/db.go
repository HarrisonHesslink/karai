package database

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	config "github.com/karai/go-karai/configuration"
	"strconv"
	"log"
)

type Database struct {
	Cf *config.Config
	thisSubgraph          string 
	thisSubgraphShortName string 
	poolInterval          int  
	txCount               int   
}

// Graph is a collection of transactions
type Graph struct {
	Transactions []transaction.Transaction `json:"transactions"`
}

// connect will create an active DB connection
func (d Database) Connect() (*sqlx.DB, error) {
	connectParams := fmt.Sprintf("user=%s host=localhost port=%s dbname=%s sslmode=%s password=%s", d.Cf.GetDBUser(), d.Cf.DbPort, d.Cf.GetDBName(), d.Cf.GetDBSSL(), d.Cf.DbPassword)
	db, err := sqlx.Connect("postgres", connectParams)
	util.Handle("Error creating a DB connection: ", err)
	return db, err
}

func (d *Database) DB_init() {
	d.CreateTables()
	d.CreateRoot()
}

func (d *Database) CreateTables() {
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	tx := db.MustBegin()

	q := "CREATE TABLE IF NOT EXISTS " + d.Cf.GetTableName() + "(tx_time CHAR(19) NOT NULL, tx_type CHAR(1) NOT NULL, tx_hash CHAR(128) NOT NULL, tx_data TEXT NOT NULL, tx_prev CHAR(128) NOT NULL, tx_epoc TEXT NOT NULL, tx_subg CHAR(128) NOT NULL, tx_prnt CHAR(128), tx_mile BOOLEAN NOT NULL,tx_lead BOOLEAN NOT NULL);"
	tx.MustExec(q)
	tx.Commit()
}


// createRoot Transaction channels start with a rootTx transaction always
func (d Database) CreateRoot() error {
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + d.Cf.GetTableName() + " ORDER BY $1 DESC", "tx_time").Scan(&count)
	switch {
	case err != nil:
		util.Handle("There was a problem counting database transactions: ", err)
		return err
	default:
		//  fmt.Printf("Found %v transactions in the db.", count)
		if count == 0 {
			txTime := "1603203489229912200"
			txType := "1"
			txSubg := "0"
			txPrnt := "0"
			txData := "8d3729b91a13878508c564fbf410ae4f33fcb4cfdb99677f4b23d4c4adb447650964b4fe9da16299831b9cc17aaabd5b8d81fb05460be92af99d128584101a30" // ?
			txPrev := "c66f4851618cd53104d4a395212958abf88d96962c0c298a0c7a7c1242fac5c2ee616c8c4f140a2e199558ead6d18ae263b2311b590b0d7bf3777be5b3623d9c" // RockSteady was here
			hash := sha512.Sum512([]byte(txTime + txType + txData + txPrev))
			txHash := hex.EncodeToString(hash[:])
			txMile := true
			txLead := false
			txEpoc := txHash
			tx := db.MustBegin()
			tx.MustExec("INSERT INTO " + d.Cf.GetTableName() + " (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", txTime, txType, txHash, txData, txPrev, txEpoc, txSubg, txPrnt, txMile, txLead)
			tx.Commit()
			return nil
		} else if count > 0 {
			return errors.New("Root tx already present. ")
		}
	}
	return nil
}

func (d Database) CommitDBTx(tx transaction.Transaction) {
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	txn := db.MustBegin()

	txn.MustExec("INSERT INTO " + d.Cf.GetTableName() + " (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)", tx.Time, tx.Type, tx.Hash, tx.Data, tx.Prev, tx.Epoc, tx.Subg, tx.Prnt, tx.Mile, tx.Lead)
	txn.Commit()
	d.txCount++
}

func (d Database) GetPrevHash() transaction.Transaction {
	var txData string
	var txSubg string
	var txPrev string
	var txPrnt string
	var txType string
	var txLead bool

	var txTime string
	var txHash string
	var txEpoc string
	var txMile bool

	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	_ = db.QueryRow("SELECT (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead ) FROM " + d.Cf.GetTableName() + " ORDER BY tx_time DESC LIMIT 1").Scan(&txTime, &txType, &txHash, &txData, &txPrev, &txEpoc, &txSubg, &txPrnt, &txMile, &txLead)

	return transaction.Transaction{txTime, txType, txHash, txData, txPrev, txEpoc, txSubg, txPrnt, txMile, txLead}
}

func (d Database) GetDAGSize() int {
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	rows, err := db.Queryx("SELECT * FROM " + d.Cf.GetTableName())
	if err != nil {
		log.Println(err.Error())
	}
	defer rows.Close()
	x := 0
	for rows.Next() {
		x++
	}
	return x
}
func (d *Database) GetDAGSize() int {
    db, connectErr := d.Connect()
    defer db.Close()
    util.Handle("Error creating a DB connection: ", connectErr)
    rows, err := db.Queryx("SELECT * FROM " +  d.Cf.GetTableName())
    if err != nil {
        log.Println(err.Error())
    }
    defer rows.Close()
    x := 0
    for rows.Next() {
        x++
    }
    return x
}

func (d *Database) ReturnTopHash() (string, int) {
	var txHash string
	var id int
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	_ = db.QueryRow("SELECT tx_hash, id FROM " + d.Cf.GetTableName() + " WHERE tx_type='1' ORDER BY tx_time DESC LIMIT 1").Scan(&txHash, &id)
	log.Println(txHash)
	return txHash, id
}

func(d *Database) HaveTx(hash string) bool {
	exists := true
	var tx_hash string
	db, connectErr := d.Connect()
	util.Handle("Error creating a DB connection: ", connectErr)
	defer db.Close()
	row := db.QueryRow("SELECT tx_hash FROM " + d.Cf.GetTableName() + " WHERE tx_hash=$1", hash)
	_ = row.Scan(&tx_hash); 
	if tx_hash != hash {
		exists = false
	}

    return exists
}

func (d Database) ReturnRangeOfTransactions(height int) [][]byte {

	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)

	var txes [][]byte

	rows, err := db.Query("SELECT tx_hash FROM " + d.Cf.GetTableName() + " WHERE id >" + strconv.Itoa(height))
	if err != nil {
		// handle this error better than this
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var tx_hash string
		err = rows.Scan(&tx_hash)
		if err != nil {
			// handle this error
			panic(err)
		}
		
		txes = append(txes, []byte(tx_hash))
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return txes
}

func (d *Database) GetTransaction(hash []byte) transaction.Transaction {
	var txData string
	var txSubg string
	var txPrev string
	var txPrnt string
	var txType string
	var txLead bool

	var txTime string
	var txHash string
	var txEpoc string
	var txMile bool

	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	_ = db.QueryRow("SELECT (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead) FROM " + d.Cf.GetTableName() + " WHERE tx_hash=" + string(hash) + " LIMIT 1").Scan(&txTime, &txType, &txHash, &txData, &txPrev, &txEpoc, &txSubg, &txPrnt, &txMile, &txLead)

	return transaction.Transaction{txTime, txType, txHash, txData, txPrev, txEpoc, txSubg, txPrnt, txMile, txLead}
}

// newSubGraphTimer timer for collection interval
func (d Database) newSubGraphTimer() {

	// fmt.Printf(Brightcyan+"\nSubgraph created:"+Brightgreen+" %s.."+Brightcyan+" SubGraph Interval: "+Brightgreen+"%vs\n"+nc, thisSubgraph[0:8], poolInterval)
	time.Sleep(time.Duration(d.poolInterval) * time.Second)
	d.txCount = 0
	// fmt.Printf(Brightyellow + "\nInterval concluded" + nc)
}

