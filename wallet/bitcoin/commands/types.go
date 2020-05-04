package commands

type Address string

type AddressInfo struct {
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       Address `json:"address"`
	Label         string  `json:"label,omitempty"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Confirmations int64   `json:"confirmations"`
	Spendable     bool    `json:"spendable"`
	Solvable      bool    `json:"solvable"`
	Desc          string  `json:"desc,omitempty"`
	Safe          bool    `json:"safe"`
}
