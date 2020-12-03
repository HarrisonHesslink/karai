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

type NewBlock struct {
	Height   int64      `json:"height"`
	Pubkey   string     `json:"pubkey"`
	Nodes    []string   `json:"nodes"`
	Leader   bool       `json:"leader"`
	Swaps    [][]string `json:"swaps"`
	Requests [][]string `json:"requests"`
}

type OracleData struct {
	Hash      string  `json:"hash"`
	Price     float64 `json:"price"`
	Height    int64   `json:"height"`
	Contract  string  `json:"contract"`
	Pubkey    string  `json:"pubkey"`
	Signature string  `json:"signature"`
}

type Request_Consensus struct {
	Hash      string   `json:"hash"`
	PubKey    string   `json:"pubkey"`
	Signature string   `json:"signature"`
	Data      []string `json:"data"`
	Task      string   `json:"task"`
	Height    int64    `json:"height"`
}

// Transaction This is the structure of the transaction
type Transaction struct {
	Time   string `json:"time" db:"tx_time"`
	Type   string `json:"type" db:"tx_type"`
	Hash   string `json:"hash" db:"tx_hash"`
	Data   string `json:"data" db:"tx_data"`
	Prev   string `json:"prev" db:"tx_prev"`
	Epoc   string `json:"epoc" db:"tx_epoc"`
	Subg   string `json:"subg" db:"tx_subg"`
	Prnt   string `json:"prnt" db:"tx_prnt"`
	Mile   bool   `json:"mile" db:"tx_mile"`
	Lead   bool   `json:"lead" db:"tx_lead"`
	Height int64  `json:"height" db:"tx_height"`
}

type Request_Contract struct {
	Asset string `json:asset`
	Denom string `json:denom`
}

type Trusted_Data struct {
	TrustedData   []OracleData `json:"trusted_data"`
	TrustedAnswer float64      `json:trusted_answer`
}
