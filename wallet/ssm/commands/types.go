package commands

type NewAddressResponse struct {
	Chain       string `json:"chain"`
	Address     string `json:"address"`
	PubKey      string `json:"pubkey"`
	BlindingKey string `json:"blinding_key"`
}

type SsmPath struct {
	Chain       string
	Fingerprint string
	Path        string
}

type SignTxInputs struct {
	SsmPath
	Amount          float64
	ValueCommitment string
}

type SignTxResponse struct {
	Chain    string `json:"chain"`
	SignedTx string `json:"signed_tx"`
}
