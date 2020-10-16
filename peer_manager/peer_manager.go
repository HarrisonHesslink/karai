package peer_manager

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"github.com/karai/go-karai/util"
)

func (pm PeerManager) banPeer(peerPubKey string) {
	fmt.Printf("\nBanning peer: %s" + peerPubKey[:8] + "...")
	whitelist := pm.Cf.Getp2pWhitelistDir() + "/" + peerPubKey + ".cert"
	blacklist := pm.Cf.Getp2pBlacklistDir() + "/" + peerPubKey + ".cert"
	err := os.Rename(whitelist, blacklist)
	util.Handle("Error banning peer: ", err)
}

func (pm PeerManager) unBanPeer(peerPubKey string) {
	fmt.Printf("\nUnbanning peer: %s" + peerPubKey[:8] + "...")
	whitelist := pm.Cf.Getp2pWhitelistDir() + "/" + peerPubKey + ".cert"
	blacklist := pm.Cf.Getp2pBlacklistDir() + "/" + peerPubKey + ".cert"
	err := os.Rename(blacklist, whitelist)
	util.Handle("Error unbanning peer: ", err)
}

func (pm PeerManager) blackList() {
	fmt.Printf(util.Brightcyan + "\nDisplaying banned peers...")
	files, err := ioutil.ReadDir(pm.Cf.Getp2pBlacklistDir())
	util.Handle(util.Brightred+"There was a problem retrieving the blacklist: ", err)
	for _, cert := range files {
		certName := cert.Name()
		bannedPeerPubKey := strings.TrimRight(certName, ".cert")
		fmt.Printf(util.Brightred + "\n" + bannedPeerPubKey)
	}
}

func (pm PeerManager) clearBlackList() {
	fmt.Printf(util.Brightcyan + "Clearing banned peers...")
	files, err := ioutil.ReadDir(pm.Cf.Getp2pBlacklistDir())
	util.Handle(util.Brightred+"There was a problem clearing the blacklist: ", err)
	for _, cert := range files {
		certName := cert.Name()
		bannedPeerPubKey := strings.TrimRight(certName, ".cert")
		pm.unBanPeer(bannedPeerPubKey)
	}
}

func (pm PeerManager) clearPeerList() {
	fmt.Printf(util.Brightcyan + "Purging peer certificates...")
	directory := pm.Cf.Getp2pWhitelistDir() + "/"
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)
	for index := range dirFiles {
		fileHere := dirFiles[index]
		nameHere := fileHere.Name()
		fmt.Printf(util.Brightred+"Purging: %s", nameHere)
		fullPath := directory + nameHere
		os.Remove(fullPath)
	}
	fmt.Printf(util.Brightyellow + "\n" + "Peer list empty!")

}

func (pm PeerManager) whiteList() {
	fmt.Printf(util.Brightcyan + "Displaying peers...\n")
	files, err := ioutil.ReadDir(pm.Cf.Getp2pWhitelistDir())
	util.Handle(util.Brightred+"There was a problem retrieving the peerlist: ", err)
	for _, cert := range files {
		certName := cert.Name()
		peerPubKey := strings.TrimRight(certName, ".cert")
		fmt.Printf("\n" + peerPubKey)
	}
}
