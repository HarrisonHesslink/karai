package network

import (
	"fmt"
	"net/http"
	"strconv"

	"encoding/json"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
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
		response, err := json.Marshal(map[string]bool{"status":true})
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

	api.HandleFunc("/transactions/{txs}", func(w http.ResponseWriter, r *http.Request) {
		txQuery := ""
		qry := mux.Vars(r)["txs"]
		numOfTxs, err := strconv.Atoi(qry)
		if err != nil {
			txQuery = qry
			if txQuery == "all" {
				numOfTxs = 1000000000
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		db, err := s.Prtl.Dat.Connect()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		defer db.Close()

		var queryExtension string
		if txQuery != "" && txQuery != "all" {
			queryExtension = fmt.Sprintf(` WHERE tx_hash = '%s'`, txQuery)
		}

		var transactions []transaction.Transaction
		rows, _ := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + queryExtension)
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
		var req transaction.Request_Data_TX
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println("We are data boy")

		go s.NewDataTxFromCore(req)
	}).Methods("POST")

	api.HandleFunc("/get_contracts", func(w http.ResponseWriter, r *http.Request) {

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
	}).Methods("GET")

	api.HandleFunc("/new_consensus_tx", func(w http.ResponseWriter, r *http.Request) {
		var req transaction.Request_Consensus_TX
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println("We are consensus man")

		go s.NewConsensusTXFromCore(req)
	}).Methods("POST")

	// Serve via HTTP
	log.Println("TX API listening on [::]:4203")
	http.ListenAndServe(":4203", handlers.CORS(headersCORS, originsCORS, methodsCORS)(api))
}

func (s *Server) reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}
