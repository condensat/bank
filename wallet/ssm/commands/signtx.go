package commands

import (
	"context"
	"encoding/json"
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
		if len(input.ValueCommitment) == 0 {
			amounts = fmt.Sprintf("%s %.8f", amounts, utils.ToFixed(input.Amount, 8))
		} else {
			amounts = fmt.Sprintf("%s %s", amounts, input.ValueCommitment)
		}
	}
	fingerprints = strings.Trim(fingerprints, " ")
	paths = strings.Trim(paths, " ")
	amounts = strings.Trim(amounts, " ")

	var signedTx SignTxResponse
	err := callCommand(rpcClient, CmdSignTx, &signedTx, chain, inputransaction, fingerprints, paths, amounts)
	if err != nil {
		return SignTxResponse{}, err
	}

	type DebugSignTx struct {
		Fingerprints string
		Paths        string
		Amounts      string
	}

	debug := DebugSignTx{
		Fingerprints: fingerprints,
		Paths:        paths,
		Amounts:      amounts,
	}

	str, _ := json.Marshal(&debug)

	signedTx.Debug = string(str)

	return signedTx, nil
}
