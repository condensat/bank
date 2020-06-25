package commands

import (
	"context"
)

func CreateRawTransaction(ctx context.Context, rpcClient RpcClient, inputs []UTXOInfo, outputs []SpendInfo) (Transaction, error) {
	if inputs == nil {
		inputs = make([]UTXOInfo, 0)
	}

	data := make(map[string]float64)
	for _, output := range outputs {
		data[output.Address] = output.Amount
	}
	var result Transaction
	err := callCommand(rpcClient, CmdCreateRawTransaction, &result, inputs, data)
	if err != nil {
		return "", err
	}

	return result, nil
}
