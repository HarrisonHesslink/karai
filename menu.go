package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/harrisonhesslink/pythia/network"
)

// inputHandler This is a basic input loop that listens for
// a few words that correspond to functions in the app. When
// a command isn't understood, it displays the help menu and
// returns to listening to input.
func inputHandler(s *network.Server /*keyCollection *ED25519Keys*/) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("#> " + "\n")

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if strings.Compare("help", text) == 0 {
			//	menu()
		} else if strings.Compare("dag", text) == 0 {
			count := s.P2p.Database.GetDAGSize()
			log.Println("Transactions on Network: " + strconv.Itoa(count))
		} else if strings.Compare("mempool", text) == 0 {
			s.P2p.Mempool.PrintMemPool()
		} else if strings.Compare("create_contract", text) == 0 {
			go s.P2p.CreateContract()
		}
	}
}
