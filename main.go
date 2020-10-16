package main
import (
	"github.com/karai/go-karai/network"
	config "github.com/karai/go-karai/configuration"
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
	//go restAPI(keys)
	//ascii()
	go network.Protocol_Init(&c)
	//go getDataCovid19(1000)
	//go getDataOgre(500)
	//go generateRandomTransactions()
	inputHandler()
}
