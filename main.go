package main

import (
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/network"
)

// Hello Karai
func main() {
	//osCheck()
	c := config.Config_Init()
	flags(&c)
	//checkDirs()
	//createTables()
	//cleanData()
	//keys := initKeys()
	//createRoot()
	//ascii()
	var s network.Server
	go network.Protocol_Init(&c, &s)
	//go getDataCovid19(1000)
	//go getDataOgre(500)
	//go generateRandomTransactions()
	inputHandler(&s)

}
