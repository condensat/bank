package monitor

type GeneralLoad struct {
	Num1  string `json:"1"`
	Num5  string `json:"5"`
	Num15 string `json:"15"`
}

type GeneralUptime struct {
	Day     int `json:"day"`
	Hour    int `json:"hour"`
	Minutes int `json:"minutes"`
	Second  int `json:"second"`
}

type GeneralInfo struct {
	Now    string        `json:"now"`
	Load   GeneralLoad   `json:"load"`
	Uptime GeneralUptime `json:"uptime"`
}

type CointVersion struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
}

type CointStatus struct {
	Running       bool   `json:"running"`
	Bestblockhash string `json:"bestblockhash"`
	Block         int    `json:"block"`
	Headers       int    `json:"headers"`
}

type CoinInfo struct {
	Name    string       `json:"name"`
	Version CointVersion `json:"version"`
	Status  CointStatus  `json:"status"`
}

type NodeInfo struct {
	Name    string      `json:"name"`
	Coins   []CoinInfo  `json:"coins"`
	General GeneralInfo `json:"general"`
}

type NodesReport struct {
	Nodes []NodeInfo `json:"nodes"`
}
