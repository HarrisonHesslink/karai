package transaction

type Request_Data_TX struct {
	Hash      string `json:"hash"`
	PubKey    string `json:"pub_key"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
	Task      string `json:"task"`
	Height    string `json:"height"`
	Source    string `json:"source"`
	Epoc      string `json:"epoc"`
}

type Request_Consensus_TX struct {
	Hash      string   `json:"hash"`
	PubKey    string   `json:"pub_key"`
	Signature string   `json:"signature"`
	Data      []string `json:"data"`
	Task      string   `json:"task"`
	Height    string   `json:"height"`
}

// Transaction This is the structure of the transaction
type Transaction struct {
	Time string `json:"time" database:"tx_time"`
	Type string `json:"type" database:"tx_type"`
	Hash string `json:"hash" database:"tx_hash"`
	Data string `json:"data" database:"tx_data"`
	Prev string `json:"prev" database:"tx_prev"`
	Epoc string `json:"epoc" database:"tx_epoc"`
	Subg string `json:"subg" database:"tx_subg"`
	Prnt string `json:"prnt" database:"tx_prnt"`
	Mile bool   `json:"mile" database:"tx_mile"`
	Lead bool   `json:"lead" database:"tx_lead"`
}
