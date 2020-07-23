package commands

type NewAddressResponse struct {
	Address string `json:"address"`
	Chain   string `json:"chain"`
	PubKey  string `json:"pubkey"`
}
