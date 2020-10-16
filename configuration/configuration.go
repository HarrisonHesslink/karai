package configuration

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	//"io/ioutil"
	"log"
	"os"
)

type Config struct{
	AppName string `json:"app_name"`
	AppDev string `json:"app_dev"`
	AppLicense string `json:"app_license"`
	AppRepository string `json:"app_repository"`
	AppURL string `json:"app_url"`
	ConfigDir string `json:"config_dir"`
	P2pConfigDir string `json:"p2p_config_dir"`
	P2pWhitelistDir string `json:"p2p_whitelist_dir"`
	P2pBlacklistDir string `json:"p2p_blacklist_dir"`
	CertPathDir string `json:"cert_path_dir"`
	CertPathSelfDir string `json:"cert_path_self_dir"`
	CertPathRemote string `json:"cert_path_remote"`
	PubKeyFilePath string `json:"pubkey_filepath"`
	PrivKeyFilePath string `json:"privkey_filepath"`
	SignedKeyFilePath string `json:"signedkey_filepath"`
	SelfCertFilePath string `json:"self_cert_filepath"`
	PeeridPath string `json:"peerid_path"`
	ChannelName string `json:"channel_name"`
	ChannelDesc string `json:"channel_desc"`
	ChannelCont string `json:"channel_cont"`
	NodeKey string `json:"node_key"`
	DbUser string `json:"db_user"`
	DbName string `json:"db_name"`
	DbSSL string `json:"db_SSL"`
	KaraiAPIPort int `json:"karai_api_port"`
	TableName string `json:"table_name"`
	Lport int `json:"listen_port"`
	WantsClean bool `json:"wants_clean"`
}

// Attribution constants
const (
	appName        = "go-karai"
	appDev         = "The TurtleCoin Developers"
	appDescription = appName + " is the Go implementation of the Karai network spec. Karai is a universal blockchain scaling solution for distributed applications."
	appLicense     = "https://choosealicense.com/licenses/mit/"
	appRepository  = "https://github.com/karai/go-karai"
	appURL         = "https://karai.io"
)

// File & folder constants
var (
	configDir         string = ""
	p2pConfigDir      = configDir + "/p2p"
	p2pWhitelistDir   = p2pConfigDir + "/whitelist"
	p2pBlacklistDir   = p2pConfigDir + "/blacklist"
	certPathDir       = p2pConfigDir + "/cert"
	certPathSelfDir   = certPathDir + "/self"
	certPathRemote    = certPathDir + "/remote"
	pubKeyFilePath    = certPathSelfDir + "/" + "pub.key"
	privKeyFilePath   = certPathSelfDir + "/" + "priv.key"
	signedKeyFilePath = certPathSelfDir + "/" + "signed.key"
	selfCertFilePath  = certPathSelfDir + "/" + "self.cert"
	peeridPath    = certPathSelfDir + "/" + "my_peerid.json"

)

// Channel values
const (
	channelName string = "⏣ Equilibria"
	channelDesc string = "This is an Equilibra task channel."
	channelCont string = "harrison@equilibria.network"
	nodeKey string = "TvzBcga1Kc1GJriZmNdXSR7axVd87k34kfmEtoN1Ua77iNyTKp8c3jBU7qfSRSATjoRVutoG87bD62jfY3F8AdCK1ETzHQk4B"
)

// Coordinator values
var (
	dbUser       string = "postgres"
	dbName       string = "postgres"
	dbSSL        string = "disable"
	joinMsg      []byte = []byte("JOIN")
	ncasMsg      []byte = []byte("NCAS")
	capkMsg      []byte = []byte("CAPK")
	certMsg      []byte = []byte("CERT")
	peerMsg      []byte = []byte("PEER")
	pubkMsg      []byte = []byte("PUBK")
	nsigMsg      []byte = []byte("NSIG")
	sendMsg      []byte = []byte("SEND")
	rtrnMsg      []byte = []byte("RTRN")
	numTx        int
	wantsClean   bool = false
	karaiAPIPort int
	upgrader     = websocket.Upgrader{
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
	}
)

// Matrix Values
var (
	wantsMatrix  bool   = false
	wantsFiles   bool   = true
	matrixToken  string = ""
	matrixURL    string = ""
	matrixRoomID string = ""
)

// Subgraph values
var (
	thisSubgraph          string = ""
	thisSubgraphShortName string = ""
	poolInterval          int    = 10 // seconds
	txCount               int    = 0
)

var (
	port int = 42001
	target string = ""
	insecure bool = true
	seed int = 1241241254
	db_name string = ""
)

var bootstrapPeers = []string{
	"",
	"",
}

func Config_Init() Config {
	var con Config
	r, err := os.Open("./config.json")
	if err != nil {
		log.Panic(err)
	}
	decoder := json.NewDecoder(r)
	err = decoder.Decode(&con)
	if err != nil {
		log.Panic(err)
	}

	return con
}

func (c Config) GetAppName() string {
	return c.AppName
}

func (c Config) GetAppDev() string {
	return c.AppDev
}

func (c Config) GetConfigDir() string {
	return c.ConfigDir
}

func (c Config) Getp2pConfigDir() string {
	return c.GetConfigDir() + c.P2pConfigDir
}

func (c Config) Getp2pWhitelistDir() string {
	return c.Getp2pConfigDir() + c.P2pWhitelistDir
}

func (c Config) Getp2pBlacklistDir() string {
	return c.Getp2pConfigDir() + c.P2pBlacklistDir
}

func (c Config) GetcertPathDir() string {
	return c.Getp2pConfigDir() + c.CertPathDir
}

func (c Config) GetcertPathSelfDir() string {
	return c.GetcertPathDir() + c.CertPathSelfDir
}

func (c Config) GetcertPathRemote() string {
	return c.GetcertPathDir() + c.CertPathRemote
}

func (c Config) GetWantsClean() bool {
	return c.WantsClean
}

func (c Config) GetDBName() string {
	return c.DbName
}

func (c Config) GetDBUser() string {
	return c.DbUser
}

func (c Config) GetDBSSL() string {
	return c.DbSSL
}

func (c Config) GetTableName() string {
	return c.TableName
}

func (c Config) GetListenPort() int {
	return c.Lport
}


func (c Config) SetkaraiAPIPort(port int) {
	c.Lport = port
}

func (c Config) SetwantsClean(wants_clean bool) {
	c.WantsClean = wants_clean
}

func (c Config) SetconfigDir(dir string) {
	c.ConfigDir = dir
}

func (c Config) Setlport(port int) {
	c.Lport = port
}

func (c Config) SettableName(name string) {
	c.TableName = name
}



