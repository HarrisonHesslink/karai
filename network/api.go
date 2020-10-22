package network

import (
	"net/http"
	//"strconv"
	"log"
	 "github.com/gorilla/handlers"
	 "github.com/gorilla/mux"
	"github.com/karai/go-karai/transaction"
	"github.com/karai/go-karai/util"
	"encoding/json"
	"github.com/gorilla/websocket"
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

	api.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		// upgrade this connection to a WebSocket
		// connection
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
		}
		db, connectErr := s.Prtl.Dat.Connect()
		defer db.Close()
		util.Handle("Error creating a DB connection: ", connectErr)

		row3, err := db.Queryx("SELECT * FROM " + s.Prtl.Dat.Cf.GetTableName() + " ORDER BY tx_time ASC")
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

			json_string, _ := json.Marshal(t2_tx)

			err = ws.WriteMessage(1, json_string)
			if err != nil {
				log.Println(err)
			}
		}
		// listen indefinitely for new messages coming
		// through on our WebSocket connection
		go s.reader(ws)

	})

	

	// Home
	//api.HandleFunc("/", home).Methods(http.MethodGet)

	// Version
	//api.HandleFunc("/version", returnVersion).Methods(http.MethodGet)

	// Stats
	// api.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
	// 	returnStatsWeb(w, r, keyCollection)
	// }).Methods(http.MethodGet)

	// Transaction by ID
	// api.HandleFunc("/transaction/{hash}", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	hash := vars["hash"]
	// 	returnSingleTransaction(w, r, hash)
	// }).Methods(http.MethodGet)

	// Transaction by qty
	// api.HandleFunc("/transactions/{number}", func(w http.ResponseWriter, r *http.Request) {
	// 	vars := mux.Vars(r)
	// 	number := vars["number"]
	// 	returnNTransactions(w, r, number)
	// }).Methods(http.MethodGet)

	api.HandleFunc("/new_tx", func(w http.ResponseWriter, r *http.Request) {
		var req transaction.Request_Data_TX
		// Try to decode the request body into the struct. If there is an error,
		// respond to the client with the error message and a 400 status code.
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
		
		json,_ := json.Marshal(ArrayTX{transactions})

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

	// api.HandleFunc("/tx_api", func(w http.ResponseWriter, r *http.Request) {
	// 	var upgrader = websocket.Upgrader{}
	// 	conn, _ := upgrader.Upgrade(w, r, nil)
	// 	defer conn.Close()
	// 	log.Println("socket open")
	// 	s.HandleAPISocket(conn)
	// })

	// Serve via HTTP
	http.ListenAndServe(":4203", handlers.CORS(headersCORS, originsCORS, methodsCORS)(api))
}

func (s * Server) reader(conn *websocket.Conn) {
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