package database

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"

	config "github.com/harrisonhesslink/pythia/configuration"
	"github.com/harrisonhesslink/pythia/transaction"
	"github.com/harrisonhesslink/pythia/util"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Database struct {
	Cf                    *config.Config
	thisSubgraph          string
	thisSubgraphShortName string
	poolInterval          int
	txCount               int

	Mutex sync.Mutex
}

// Graph is a collection of transactions
type Graph struct {
	Transactions []transaction.Transaction `json:"transactions"`
}

func NewDataBase(c *config.Config) *Database {
	d := new(Database)
	d.Cf = c

	d.DB_init()

	return d
}

// connect will create an active DB connection
func (d *Database) Connect() (*sqlx.DB, error) {
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

	q := "CREATE TABLE IF NOT EXISTS " + d.Cf.GetTableName() + "(tx_time CHAR(19) NOT NULL, tx_type CHAR(1) NOT NULL, tx_hash CHAR(128) NOT NULL, tx_data TEXT NOT NULL, tx_prev CHAR(128) NOT NULL, tx_epoc TEXT NOT NULL, tx_subg CHAR(128) NOT NULL, tx_prnt CHAR(128), tx_mile BOOLEAN NOT NULL,tx_lead BOOLEAN NOT NULL, tx_height INTEGER NOT NULL);"
	tx.MustExec(q)
	tx.Commit()
}

// createRoot Transaction channels start with a rootTx transaction always
func (d Database) CreateRoot() error {
	db, connectErr := d.Connect()
	defer db.Close()
	util.Handle("Error creating a DB connection: ", connectErr)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM "+d.Cf.GetTableName()+" ORDER BY $1 DESC", "tx_time").Scan(&count)
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
			tx.MustExec("INSERT INTO "+d.Cf.GetTableName()+" (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead, tx_height ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", txTime, txType, txHash, txData, txPrev, txEpoc, txSubg, txPrnt, txMile, txLead, 0)
			tx.Commit()
			return nil
		} else if count > 0 {
			return errors.New("Root tx already present. ")
		}
	}
	return nil
}

func (d *Database) CommitDBTx(tx transaction.Transaction) {

	if !d.HaveTx(tx.Hash) {
		db, connectErr := d.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)
		txn := db.MustBegin()

		txn.MustExec("INSERT INTO "+d.Cf.GetTableName()+" (tx_time, tx_type, tx_hash, tx_data, tx_prev, tx_epoc, tx_subg, tx_prnt, tx_mile, tx_lead, tx_height ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", tx.Time, tx.Type, tx.Hash, tx.Data, tx.Prev, tx.Epoc, tx.Subg, tx.Prnt, tx.Mile, tx.Lead, tx.Height)
		txn.Commit()
	}
}

func (d *Database) GetDAGSize() int {
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

func (d *Database) HaveTx(hash string) bool {
	db, connectErr := d.Connect()
	util.Handle("Error creating a DB connection: ", connectErr)
	defer db.Close()

	var exists bool
	err := db.QueryRow("SELECT exists(select 1 from transactions where tx_hash=$1)", hash).Scan(&exists)
	if err != nil {
		return true
	}
	return exists
}
