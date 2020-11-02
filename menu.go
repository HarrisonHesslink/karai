package main

import (
	"bufio"
	"fmt"
	"github.com/karai/go-karai/network"
	"log"
	"os"
	"strconv"
	"strings"
)

// inputHandler This is a basic input loop that listens for
// a few words that correspond to functions in the app. When
// a command isn't understood, it displays the help menu and
// returns to listening to input.
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
		} else if strings.Compare("mempool", text) == 0 {
			s.Prtl.Mempool.PrintMemPool()
		} else if strings.HasPrefix(text, "create_contract ") {
			strings.TrimPrefix(text, "create_contract ")
			args := strings.Fields(text)
			if args[1] == "XHV" || args[1] == "XEQ" || args[1] == "LOKI" || args[1] == "ETH" || args[1] == "DOGE" || args[1] == "XMR" {
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

// // func inputHandler(keyCollection *ED25519Keys) {
// // 	reader := bufio.NewReader(os.Stdin)
// // 	var conn *websocket.Conn

// // 	fmt.Printf("\n\n%v%v%v\n", white+"Type '", brightgreen+"menu", white+"' to view a list of commands")
// // 	for {
// // 		// if isCoordinator {
// // 		fmt.Print(brightred + "#> " + nc)
// // 		// }
// // 		// if !isCoordinator {
// // 		// 	fmt.Print(brightgreen + "$> " + nc)
// // 		// }
// // 		text, _ := reader.ReadString('\n')
// // 		text = strings.TrimSpace(text)
// // 		if strings.Compare("help", text) == 0 {
// // 			menu()
// // 		} else if strings.Compare("?", text) == 0 {
// // 			menu()
// // 		} else if strings.Compare("peer", text) == 0 {
// // 			fmt.Printf(brightcyan + "Peer ID: ")
// // 			fmt.Printf(cyan+"%s\n", keyCollection.publicKey)
// // 		} else if strings.Compare("menu", text) == 0 {
// // 			menu()
// // 		} else if strings.Compare("version", text) == 0 {
// // 			menuVersion()
// // 		} else if strings.Compare("license", text) == 0 {
// // 			printLicense()
// // 		} else if strings.Compare("upgrade", text) == 0 {
// // 			returnMessage(conn, keyCollection.publicKey)
// // 		} else if strings.Compare("a", text) == 0 {
// // 			// if isCoordinator {
// // 			start := time.Now()
// // 			txint := 50
// // 			addBulkTransactions(txint)

// // 			elapsed := time.Since(start)
// // 			fmt.Printf("\nWriting %v objects to memory took %s seconds.\n", txint, elapsed)
// // 			// } else {
// // 			// 	fmt.Printf("It looks like you're not a channel coordinator. \nRun Karai with '-coordinator' option to run this command.\n")
// // 			// }
// // 		} else if strings.HasPrefix(text, "connect") {
// // 			ktxAddressString := strings.TrimPrefix(text, "connect ")
// // 			if strings.Contains(ktxAddressString, ":") {
// // 				var justTheDomainPartNotThePort = strings.Split(ktxAddressString, ":")
// // 				var ktxCertFileName = certPathDir + "/remote/" + justTheDomainPartNotThePort[0] + ".cert"
// // 				if !fileExists(ktxCertFileName) {
// // 					joinChannel(ktxAddressString, keyCollection.publicKey, keyCollection.signedKey, "", keyCollection)
// // 				}
// // 				if fileExists(ktxCertFileName) {
// // 					isFNG = false
// // 					joinChannel(ktxAddressString, keyCollection.publicKey, keyCollection.signedKey, ktxCertFileName, keyCollection)
// // 				}

// // 			}
// // 			if !strings.Contains(ktxAddressString, ":") {
// // 				fmt.Printf("\nDid you forget to include the port?\n")
// // 			}
// // 		} else if strings.HasPrefix(text, "ban ") {
// // 			bannedPeer := strings.TrimPrefix(text, "ban ")
// // 			banPeer(bannedPeer)
// // 		} else if strings.HasPrefix(text, "unban ") {
// // 			unBannedPeer := strings.TrimPrefix(text, "unban ")
// // 			unBanPeer(unBannedPeer)
// // 		} else if strings.HasPrefix(text, "blacklist") {
// // 			blackList()
// // 		} else if strings.Compare("clear blacklist", text) == 0 {
// // 			clearBlackList()
// // 		} else if strings.Compare("clear peerlist", text) == 0 {
// // 			clearPeerList()
// // 		} else if strings.Compare("peerlist", text) == 0 {
// // 			whiteList()
// // 		} else if strings.Compare("exit", text) == 0 {
// // 			menuExit()
// // 		} else if strings.Compare("create-channel", text) == 0 {
// // 			fmt.Printf(brightcyan + "Creating channel...\n" + white)
// // 			// spawnChannel()
// // 		} else if strings.Compare("generate-pointer", text) == 0 {
// // 			generatePointer()
// // 		} else if strings.Compare("quit", text) == 0 {
// // 			menuExit()
// // 		} else if strings.Compare("close", text) == 0 {
// // 			menuExit()
// // 		}
// // 	}
// // }

// func menu() {
// 	menuOptions := []string{"LAUNCH_PARAMETERS", "CHANNEL_OPTIONS", "WALLET_API_OPTIONS", "KARAI_OPTIONS", "GENERAL_OPTIONS"}
// 	menuData := map[string][][]string{
// 		"LAUNCH_PARAMETERS": {
// 			{
// 				"-https \t\t\t Use HTTPS for Coordinator API",
// 				"-matrix \t\t Send event messages to Matrix homeserver",
// 				"-matrixtoken \t\t Matrix homeserver token string",
// 				"-matrixurl \t\t Matrix homeserver URL string",
// 				"-matrixroomid \t\t Room ID string for matrix publishd events",
// 				"-apiport \t\t Coordinator API port integer",
// 			},
// 			{},
// 		},
// 		"CHANNEL_OPTIONS": {
// 			{
// 				"generate-pointer \t Generate a Karai <=> TRTL pointer",
// 			},
// 			{},
// 		},
// 		"WALLET_API_OPTIONS": {
// 			{},
// 			{
// 				// "open-wallet \t\t Open a TRTL wallet",
// 				// "open-wallet-info \t Show wallet and connection info",
// 				// "create-wallet \t\t Create a TRTL wallet",
// 				// "wallet-balance \t\t Displays wallet balance",
// 			},
// 		},
// 		"KARAI_OPTIONS": {
// 			{
// 				"peerlist \t\t Lists known peers.",
// 				"blacklist \t\t Lists banned peers.",
// 				"clear blacklist \t Unbans all blacklist peer certificates.",
// 				"clear peerlist \t\t Purges all whitelist peer certificates.",
// 				"ban <pubkey> \t\t Ban user certificate by pubkey.",
// 				"unban <pubkey> \t\t Unban user certificate by pubkey.",
// 			},
// 		},
// 		"GENERAL_OPTIONS": {
// 			{
// 				"version \t\t Displays version",
// 				"license \t\t Displays license",
// 				"exit \t\t\t Quit immediately",
// 			},
// 		},
// 	}

// 	for _, opt := range menuOptions {
// 		fmt.Printf(brightgreen + "\n" + opt)
// 		for menuOptionColor, options := range menuData[opt] {
// 			switch menuOptionColor {
// 			case 0:
// 				fmt.Printf(brightwhite)
// 			case 1:
// 				fmt.Printf(brightblack)
// 			}
// 			for _, message := range options {
// 				fmt.Printf("\n" + message)
// 			}
// 		}
// 	}

// 	fmt.Printf("\n")
// }
