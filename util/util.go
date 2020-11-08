package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	config "github.com/karai/go-karai/configuration"
)

// ascii Splash logo. We used to have a package for this
// but it was only for the logo so why not just static-print it?
func ascii() {
	fmt.Printf("\n\n")
	fmt.Printf(green + "|  |/  / /  /\\  \\ |  |)  ) /  /\\  \\ |  |\n")
	fmt.Printf(Brightgreen + "|__|\\__\\/__/¯¯\\__\\|__|\\__\\/__/¯¯\\__\\|__| \n")
	fmt.Printf(Brightred + "v" + semverInfo() + white)
	fmt.Printf(Brightred + " coordinator")

}

// StatsDetail is an object containing strings relevant to the status of a coordinator node.
type StatsDetail struct {
	ChannelName        string `json:"channel_name"`
	ChannelDescription string `json:"channel_description"`
	Version            string `json:"version"`
	ChannelContact     string `json:"channel_contact"`
	PubKeyString       string `json:"pub_key_string"`
	TxObjectsOnDisk    int    `json:"tx_objects_on_disk"`
	GraphUsers         int    `json:"tx_graph_users"`
}

func delay(seconds time.Duration) {
	time.Sleep(seconds * time.Second)
}

// printLicense Print the license for the user
func printLicense(c *config.Config) {
	fmt.Printf(Brightgreen + "\n" + c.GetAppName() + " v" + semverInfo() + white + " by " + c.GetAppDev())
	fmt.Printf(Brightgreen + "\n" + "s" + "\n" + c.GetAppDev() + "\n")
	fmt.Printf(Brightwhite + "\nMIT License\nCopyright (c) 2020-2021 RockSteady, TurtleCoin Developers")
	fmt.Printf(Brightblack + "\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in allcopies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.")
	fmt.Printf("\n")
}

// menuVersion Print the version string for the user
func menuVersion(c *config.Config) {
	fmt.Printf("%s - v%s\n", c.GetAppName(), semverInfo())
}

// fileExists Does this file exist?
func fileExists(filename string) bool {
	referencedFile, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !referencedFile.IsDir()
}

// fileContainsString This is a utility to see if a string in a file exists.
func fileContainsString(str, filepath string) bool {
	accused, _ := ioutil.ReadFile(filepath)
	isExist, _ := regexp.Match(str, accused)
	return isExist
}

// menuExit Exit the program
func menuExit() {
	os.Exit(0)
}

func timeStamp() string {
	current := time.Now()
	return current.Format("2006-01-02 15:04:05")
}

func UnixTimeStampNano() string {
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	return timestamp
}

func writeTxToDisk(gtxType, gtxHash, gtxData, gtxPrev string) {
	timeNano := UnixTimeStampNano()
	txFileName := timeNano + ".json"
	createFile(txFileName)
	txJSONItems := []string{gtxType, gtxHash, gtxData, gtxPrev}
	txJSONObject, _ := json.Marshal(txJSONItems)
	fmt.Printf(white+"\nWriting file...\nFileName: %s\nTransaction Body Object\n%s", txFileName, string(txJSONObject))
	writeFile(txFileName, string(txJSONObject))
}
func createDirIfItDontExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		Handle("Could not create directory: ", err)
	}
}

// checkDirs Check if directory exists
func checkDirs(c *config.Config) {
	fmt.Printf("\n" + c.GetConfigDir())
	createDirIfItDontExist(c.GetConfigDir())
	createDirIfItDontExist(c.Getp2pConfigDir())
	createDirIfItDontExist(c.Getp2pWhitelistDir())
	createDirIfItDontExist(c.Getp2pBlacklistDir())
	createDirIfItDontExist(c.GetcertPathDir())
	createDirIfItDontExist(c.GetcertPathSelfDir())
	createDirIfItDontExist(c.GetcertPathRemote())
}

// Handle Ye Olde Error Handler takes a message and an error code
func Handle(msg string, err error) {
	if err != nil {
		fmt.Printf(Brightred+"\n%s: %s"+white, msg, err)
	}
}

// createFile Generic file Handler
func createFile(filename string) {
	var _, err = os.Stat(filename)
	if os.IsNotExist(err) {
		var file, err = os.Create(filename)
		Handle("", err)
		defer file.Close()
	}
}

// writeFile Generic file Handler
func writeFile(filename, textToWrite string) {
	var file, err = os.OpenFile(filename, os.O_RDWR, 0644)
	Handle("", err)
	defer file.Close()
	_, err = file.WriteString(textToWrite)
	err = file.Sync()
	Handle("", err)
}

// writeFileBytes Generic file Handler
func writeFileBytes(filename string, bytesToWrite []byte) {
	var file, err = os.OpenFile(filename, os.O_RDWR, 0644)
	Handle("", err)
	defer file.Close()
	_, err = file.Write(bytesToWrite)
	err = file.Sync()
	Handle("", err)
}

// readFile Generic file Handler
func readFile(filename string) string {
	text, err := ioutil.ReadFile(filename)
	Handle("Couldnt read the file: ", err)
	return string(text)
}

func readFileBytes(filename string) []byte {
	text, err := ioutil.ReadFile(filename)
	Handle("Couldnt read the file: ", err)
	return text
}

// deleteFile Generic file Handler
func deleteFile(filename string) {
	err := os.Remove(filename)
	Handle("Problem deleting file: ", err)
}

func validJSON(stringToValidate string) bool {
	return json.Valid([]byte(stringToValidate))
}

func countWhitelistPeers(c *config.Config) int {
	directory := c.Getp2pWhitelistDir() + "/"
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)
	count := 0
	for range dirFiles {
		count++
	}
	return count
}

func cleanData(c *config.Config) {
	if c.GetWantsClean() {
		// cleanse the whitelist
		directory := c.Getp2pWhitelistDir() + "/"
		dirRead, _ := os.Open(directory)
		dirFiles, _ := dirRead.Readdir(0)
		for index := range dirFiles {
			fileHere := dirFiles[index]
			nameHere := fileHere.Name()
			fullPath := directory + nameHere
			deleteFile(fullPath)
		}

		// cleanse the blacklist
		blackList, _ := ioutil.ReadDir(c.Getp2pBlacklistDir() + "/")
		for _, f := range blackList {
			fileToDelete := c.Getp2pBlacklistDir() + "/" + f.Name()
			fmt.Printf("\nDeleting file: %s", fileToDelete)
			deleteFile(f.Name())
		}
		fmt.Printf(Brightyellow+"\nPeers clear: %s"+white, Brightgreen+"✔️")

		// cleanse the remote certs
		remoteCert, _ := ioutil.ReadDir(c.GetcertPathRemote() + "/")
		for _, f := range remoteCert {
			fileToDelete := c.GetcertPathRemote() + "/" + f.Name()
			fmt.Printf("\nDeleting file: %s", fileToDelete)
			deleteFile(fileToDelete)
		}
		fmt.Printf(Brightyellow+"\nCerts clear: %s"+white, Brightgreen+"✔️")

	}
}

//move to maybe logger package
func Success_log(msg string) {
	log.Println(Brightgreen + msg + white)
}

func Error_log(msg string) {
	log.Println(Brightred + msg + white)
}

func Warning_log(msg string) {
	log.Println(Brightyellow + msg + white)
}

func Success_log_array(msg string) {
	log.Print(Brightgreen + msg + white)
}
