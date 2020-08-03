package commands

type NewAddressResponse struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
	PubKey  string `json:"pubkey"`
}

type SsmPath struct {
	Chain       string
	Fingerprint string
	Path        string
}

type SignTxInputs struct {
	SsmPath
	Amount float64
}

type SignTxResponse struct {
	Chain    string `json:"chain"`
	SignedTx string `json:"signed_tx"`
}
