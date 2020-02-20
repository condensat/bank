package monitor

import (
	"encoding/json"
	"sort"
)

func ParseCoinInfo(name string, values interface{}) CoinInfo {
	result := CoinInfo{Name: name}

	data, err := json.Marshal(values)
	if err != nil {
		return result
	}

	err = json.Unmarshal(data, &result.Version)
	if err != nil {
		return result
	}
	err = json.Unmarshal(data, &result.Status)
	if err != nil {
		return result
	}

	return result
}

func ParseCoins(values map[string]interface{}) []CoinInfo {
	var result []CoinInfo

	for name, values := range values {
		result = append(result, ParseCoinInfo(name, values))
	}

	return result
}

func ParseNodeReport(values map[string]interface{}) NodesReport {
	var result NodesReport

	for name, values := range values {
		dataNode, err := json.Marshal(values)
		if err != nil {
			continue
		}

		var nodeData struct {
			General GeneralInfo            `json:"general"`
			Coins   map[string]interface{} `json:"coins"`
		}

		err = json.Unmarshal(dataNode, &nodeData)
		if err != nil {
			continue
		}

		result.Nodes = append(result.Nodes, NodeInfo{
			Name:    name,
			General: nodeData.General,
			Coins:   ParseCoins(nodeData.Coins),
		})

		// Sort nodes and coins
		sort.Slice(result.Nodes, func(i, j int) bool {
			return result.Nodes[i].Name < result.Nodes[j].Name
		})
		for _, node := range result.Nodes {
			sort.Slice(node.Coins, func(i, j int) bool {
				return node.Coins[i].Name < node.Coins[j].Name
			})
		}
	}

	return result
}
