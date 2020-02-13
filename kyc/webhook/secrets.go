package webhook

import (
	"encoding/json"
	"io/ioutil"
)

type Secrets struct {
	Verified string `json:"verified"`
	Failed   string `json:"failed"`
	Revoked  string `json:"revoked"`
	Ready    string `json:"ready"`
}

func FromFile(secretFile string) (Secrets, error) {
	data, err := ioutil.ReadFile(secretFile)
	if err != nil {
		return Secrets{}, err
	}

	var secrets Secrets
	err = json.Unmarshal(data, &secrets)
	if err != nil {
		return Secrets{}, err
	}

	return secrets, nil
}

func (p *Secrets) Get(name string) string {
	switch name {
	case "verified":
		return p.Verified
	case "failed":
		return p.Failed
	case "revoked":
		return p.Revoked
	case "ready":
		return p.Ready
	default:
		return ""
	}
}
