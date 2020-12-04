package main

import (
	config "github.com/harrisonhesslink/pythia/configuration"
	"github.com/harrisonhesslink/pythia/network"
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
	network.ProtocolInit(&c, &s)
	//go getDataCovid19(1000)
	//go getDataOgre(500)
	//go generateRandomTransactions()
}
