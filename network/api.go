package network

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"net/http"
	"strconv"
	"strings"
	//"strconv"
	"log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// restAPI() This is the main API that is activated when isCoord == true
func (s *Server) RestAPI() {

	// CORS
	corsAllowedHeaders := []string{
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Origin",
		"Cache-Control",
		"Content-Security-Policy",
		"Feature-Policy",
		"Referrer-Policy",
		"X-Requested-With"}

	corsOrigins := []string{
		"*",
		"127.0.0.1"}

	corsMethods := []string{
		"GET",
		"HEAD",
		"POST",
		"PUT",
		"OPTIONS"}

	headersCORS := handlers.AllowedHeaders(corsAllowedHeaders)
	originsCORS := handlers.AllowedOrigins(corsOrigins)
	methodsCORS := handlers.AllowedMethods(corsMethods)

	// Init API
	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response, err := json.Marshal(map[string]bool{"status": true})
		if err != nil {
			log.Println(err.Error())
		}
		_, _ = w.Write(response)
	})

	// Version
	//api.HandleFunc("/version", returnVersion).Methods(http.MethodGet)

	// Stats
	// api.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
	// 	returnStatsWeb(w, r, keyCollection)
	// }).Methods(http.MethodGet)

	api.HandleFunc("/transactions/{type}/{txs}", func(w http.ResponseWriter, r *http.Request) {
		var txQuery string
		qry := mux.Vars(r)["txs"]
		numOfTxs, err := strconv.Atoi(qry)
		if err != nil {
			txQuery = qry
			numOfTxs = 1000000000
		}

		_type := mux.Vars(r)["type"]
		if _type != "asc" && _type != "desc" && _type != "contract" {
			res, _ := json.Marshal(map[string]bool{"status": false})
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(res)
			return
		}

		order := " ORDER BY tx_time " + strings.ToUpper(_type)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		db, err := s.Prtl.Dat.Connect()
		if err != nil {
			res, _ := json.Marshal(map[string]bool{"status": false})
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(res)
			return
		}
		defer db.Close()

		var queryExtension string
		if txQuery != "" && txQuery != "all" {
			queryExtension = fmt.Sprintf(` WHERE tx_hash = '%s'`, txQuery)
		}
		if txQuery == "nondatatxs" {
			queryExtension = " WHERE tx_type = '1' OR tx_type = '3'"
		}
		if _type == "contract" {
			queryExtension = fmt.Sprintf(" WHERE tx_subg = '%s'", qry)
			order = ""
		}

		var transactions []transaction.Transaction
		rows, _ := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + queryExtension + order)
		defer rows.Close()
		x := 1
		for rows.Next() {
			var thisTx transaction.Transaction
			err = rows.StructScan(&thisTx)
			if err != nil {
				log.Panic(err)
			}
			transactions = append(transactions, thisTx)
			if x >= numOfTxs {
				break
			}
			x++
		}
		txs, err := json.Marshal(transactions)
		_, _ = w.Write(txs)
	}).Methods("GET")

	api.HandleFunc("/new_tx", func(w http.ResponseWriter, r *http.Request) {
		if s.Prtl.Sync.Connected == false {

			var req transaction.Request_Oracle_Data
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Println("We are data boy")
			log.Println(req)

			if (req.PubKey != "" && req.Signature != "" && req.Hash != "" && req.Task != "" && req.Data != "" && req.Height != "" && req.Source != "" && req.Epoc != "") {
				go s.NewDataTxFromCore(req)
			}
		}
		r.Body.Close()
	}).Methods("POST")

	api.HandleFunc("/get_contracts", func(w http.ResponseWriter, r *http.Request) {
		if s.Prtl.Sync.Connected == false {

			db, connectErr := s.Prtl.Dat.Connect()
			defer db.Close()
			util.Handle("Error creating a DB connection: ", connectErr)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// reportRequest("transactions/"+hash, w, r)
			transactions := []transaction.Transaction{}

			rows, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " WHERE tx_type='3' ORDER BY tx_time DESC")
			if err != nil {
				panic(err)
			}
			defer rows.Close()
			for rows.Next() {
				var thisTx transaction.Transaction
				err = rows.StructScan(&thisTx)
				if err != nil {
					// handle this error
					log.Panic(err)
				}
				transactions = append(transactions, thisTx)
			}
			// get any error encountered during iteration
			err = rows.Err()
			if err != nil {
				log.Panic(err)
			}

			json, _ := json.Marshal(ArrayTX{transactions})
			w.Write(json)
		} else {
			errorMSG := ErrorJson{"Not Done Syncing", false}
			errorJson, _ := json.Marshal(errorMSG)

			w.Write(errorJson)
		}

	}).Methods("GET")

	api.HandleFunc("/new_consensus_tx", func(w http.ResponseWriter, r *http.Request) {
		if s.Prtl.Sync.Connected == false {
			var req transaction.Request_Consensus
			// Try to decode the request body into the struct. If there is an error,
			// respond to the client with the error message and a 400 status code.
			err := json.NewDecoder(r.Body).Decode(&req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			log.Println("We are consensus man")

			if (req.PubKey != "" && req.Signature != "" && req.Hash != "" && req.Task != "" && len(req.Data) > 0 && req.Height != "") {
				go s.NewConsensusTXFromCore(req)
			}
		}
		r.Body.Close()
	}).Methods("POST")

	// Serve via HTTP
	log.Println("TX API listening on [::]:4203")
	http.ListenAndServe(":4203", handlers.CORS(headersCORS, originsCORS, methodsCORS)(api))
}