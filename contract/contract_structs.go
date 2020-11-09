package contract

//Contract Struct
type Contract struct {
	// Return variables for data
	// exchange:[]variables
	Data map[string][]string `json:"variables"`

	//Variable Rate of requests
	RequestTime int `json:"request_time"`

	//Other Contract ref i.e BTC/USD
	ContractRef string

	// % Threshold for new price
	Threshold string

	// Price Aggergate, others
	Type string

	// Output variable names
	OutputVar string

	// Api URL to request
	// exchange:url
	APIURLS map[string]string `json:"api_url"`
}
