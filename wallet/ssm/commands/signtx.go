package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"git.condensat.tech/bank/utils"
)

var (
	ErrInputsError = errors.New("Inputs errors")
)

func SignTx(ctx context.Context, rpcClient RpcClient, chain, inputransaction string, inputs ...SignTxInputs) (SignTxResponse, error) {
	if rpcClient == nil {
		return SignTxResponse{}, ErrInvalidRPCClient
	}

	if len(inputs) == 0 {
		return SignTxResponse{}, ErrInputsError
	}

	var fingerprints string
	var paths string
	var amounts string
	for _, input := range inputs {
		fingerprints = fmt.Sprintf("%s %s", fingerprints, input.Fingerprint)
		paths = fmt.Sprintf("%s %s", paths, input.Path)
		amounts = fmt.Sprintf("%s %.8f", amounts, utils.ToFixed(input.Amount, 8))
	}
	fingerprints = strings.Trim(fingerprints, " ")
	paths = strings.Trim(paths, " ")
	amounts = strings.Trim(amounts, " ")

	var signedTx SignTxResponse
	err := callCommand(rpcClient, CmdSignTx, &signedTx, chain, inputransaction, fingerprints, paths, amounts)
	if err != nil {
		return SignTxResponse{}, err
	}

	return signedTx, nil
}
