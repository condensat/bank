package commands

import (
	"context"

	"git.condensat.tech/bank/utils"
)

func CreateRawTransaction(ctx context.Context, rpcClient RpcClient, inputs []UTXOInfo, outputs []SpendInfo) (Transaction, error) {
	if inputs == nil {
		inputs = make([]UTXOInfo, 0)
	}

	// gather same address outputs
	data := make(map[string]float64)
	for _, output := range outputs {
		if _, ok := data[output.Address]; !ok {
			data[output.Address] = 0.0
		}
		data[output.Address] += output.Amount
	}

	// Fix satoshi precision
	for address, totalAmount := range data {
		data[address] = utils.ToFixed(totalAmount, 8)
	}

	var result Transaction
	err := callCommand(rpcClient, CmdCreateRawTransaction, &result, inputs, data)
	if err != nil {
		return "", err
	}

	return result, nil
}
