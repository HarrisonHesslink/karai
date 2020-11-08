package transaction

type Request_Oracle_Data struct {
	Hash      string `json:"hash"`
	PubKey    string `json:"pubkey"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
	Task      string `json:"task"`
	Height    string `json:"height"`
	Source    string `json:"source"`
	Epoc      string `json:"epoc"`
}

type Request_Consensus struct {
	Hash      string   `json:"hash"`
	PubKey    string   `json:"pubkey"`
	Signature string   `json:"signature"`
	Data      []string `json:"data"`
	Task      string   `json:"task"`
	Height    string   `json:"height"`
}

// Transaction This is the structure of the transaction
type Transaction struct {
	Time string `json:"time" db:"tx_time"`
	Type string `json:"type" db:"tx_type"`
	Hash string `json:"hash" db:"tx_hash"`
	Data string `json:"data" db:"tx_data"`
	Prev string `json:"prev" db:"tx_prev"`
	Epoc string `json:"epoc" db:"tx_epoc"`
	Subg string `json:"subg" db:"tx_subg"`
	Prnt string `json:"prnt" db:"tx_prnt"`
	Mile bool   `json:"mile" db:"tx_mile"`
	Lead bool   `json:"lead" db:"tx_lead"`
}

type Request_Contract struct {
	Asset string `json:asset`
	Denom string `json:denom`
}

type Trusted_Data struct {
	TrustedData   []Request_Oracle_Data `json:"trusted_data"`
	TrustedAnswer float32               `json:trusted_answer`
}
