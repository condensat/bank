package commands

type NewAddressResponse struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
	PubKey  string `json:"pubkey"`
}

type SignTxInputs struct {
	Fingerprint string
	Path        string
	Amount      float64
}

type SignTxResponse struct {
	Chain    string `json:"chain"`
	SignedTx string `json:"signed_tx"`
}
