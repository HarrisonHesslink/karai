package main

import (
	"bufio"
	"flag"
	"fmt"
	config "github.com/karai/go-karai/configuration"
	"github.com/karai/go-karai/network"
	"log"
	"os"
	"strconv"
	"strings"
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

func flags(c *config.Config) {
	// // flag.StringVar(&c.matrixToken, "matrixToken", "", "Matrix homeserver token")
	// // flag.StringVar(&c.matrixURL, "matrixURL", "", "Matrix homeserver URL")
	// // flag.StringVar(&c.matrixRoomID, "matrixRoomID", "", "Room ID for matrix publishd events")
	// c.SetkaraiAPIPort(flag.Int("apiport", 4200, "Port to run Karai Coordinator API on."))
	// c.SetwantsClean(flag.Bool("clean", false, "Clear all peer certs"))
	// // flag.BoolVar(&c.wantsMatrix, "matrix", false, "Enable Matrix functions. Requires -matrixToken, -matrixURL, and -matrixRoomID")
	// c.SetconfigDir(flag.String("dir", "./config", "Change the dir of all duh fyles"))
	// c.Setlport(flag.Int("l", 0, "wait for incoming connections"))

	//c.SettableName(flag.String("database-name", "transactions", "set database-name for psql"))

	apiport := flag.Int("apiport", 4200, "Port to run Karai Coordinator API on.")
	wantclean := flag.Bool("clean", false, "Clear all peer certs")
	dir := flag.String("dir", "./config", "Change the dir of all duh fyles")
	lport := flag.Int("l", 4201, "wait for incoming connections")
	name := flag.String("database-name", "transactions", "set database-name for psql")
	flag.Parse()

	c.KaraiAPIPort = *apiport
	c.WantsClean = *wantclean
	c.ConfigDir = *dir
	c.Lport = *lport
	c.TableName = *name


}

func inputHandler(s *network.Server/*keyCollection *ED25519Keys*/) {
	reader := bufio.NewReader(os.Stdin)

	//fmt.Printf("\n\n%v%v%v\n", white+"Type '", brightgreen+"menu", white+"' to view a list of commands")
	for {
		// if isCoordinator {
		fmt.Print("#> " + "\n")
		// }
		// if !isCoordinator {
		// 	fmt.Print(brightgreen + "$> " + nc)
		// }
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if strings.Compare("help", text) == 0 {
			//	menu()
		} else if strings.Compare("?", text) == 0 {
			//menu()
		} else if strings.Compare("peer", text) == 0 {
			//	fmt.Printf(brightcyan + "Peer ID: ")
			//	fmt.Printf(cyan+"%s\n", keyCollection.publicKey)
		} else if strings.Compare("menu", text) == 0 {
			//menu()
		} else if strings.Compare("version", text) == 0 {
			//menuVersion()
		} else if strings.Compare("license", text) == 0 {
			//	printLicense()
		} else if strings.Compare("dag", text) == 0 {
			count := s.Prtl.Dat.GetDAGSize()
			log.Println("Txes: " + strconv.Itoa(count))
		} else if strings.Compare("a", text) == 0 {
			// // start := time.Now()
			// // txint := 50
			// // addBulkTransactions(txint)
			// // elapsed := time.Since(start)
			// fmt.Printf("\nWriting %v objects to memory took %s seconds.\n", txint, elapsed)
		} else if strings.HasPrefix(text, "ban ") {
			// bannedPeer := strings.TrimPrefix(text, "ban ")
			// banPeer(bannedPeer)
		} else if strings.HasPrefix(text, "unban ") {
			// unBannedPeer := strings.TrimPrefix(text, "unban ")
			// unBanPeer(unBannedPeer)
		} else if strings.HasPrefix(text, "blacklist") {
			// blackList()
		} else if strings.Compare("clear blacklist", text) == 0 {
			// clearBlackList()
		} else if strings.Compare("clear peerlist", text) == 0 {
			// clearPeerList()
		} else if strings.Compare("peerlist", text) == 0 {
			// whiteList()
		} else if strings.Compare("exit", text) == 0 {
			// menuExit()
		} else if strings.Compare("generate-pointer", text) == 0 {
			// generatePointer()
		} else if strings.Compare("quit", text) == 0 {
			// menuExit()
		} else if strings.Compare("close", text) == 0 {
			// menuExit()
		} else if strings.Compare("nodes", text) == 0 {
			// nodes := "[ "
			// for _, node := range KnownNodes {
			// 	nodes += node + " "
			// }
			// nodes += "]"
			// log.Println(nodes)
		} else if strings.HasPrefix(text, "create_contract ") {
			strings.TrimPrefix(text, "create_contract ")
			args := strings.Fields(text)
			if args[1] == "XHV" || args[1] == "XEQ" || args[1] == "LOKI" || args[1] == "ETH" || args[1] == "DOGE" {
				if args[2] == "BTC" {
					go s.CreateContract(args[1], args[2])
				} else {
					log.Println("Pair Not Supported! BTC")
				}

			} else {
				log.Println("Pair Not Supported! XEQ, XHV, LOKI, ETH, DOGE")

			}
		}
	}
}
