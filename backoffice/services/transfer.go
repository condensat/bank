package services

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
)

type DepositStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type BatchStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type WithdrawStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

type TransferStatus struct {
	Deposit  DepositStatus  `json:"deposit"`
	Batch    BatchStatus    `json:"batch"`
	Withdraw WithdrawStatus `json:"withdraw"`
}

func FetchTransferStatus(ctx context.Context) (TransferStatus, error) {
	db := appcontext.Database(ctx)

	batchs, err := database.BatchsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	deposits, err := database.DepositsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	witdthdraws, err := database.WithdrawsInfos(db)
	if err != nil {
		return TransferStatus{}, err
	}

	return TransferStatus{
		Deposit: DepositStatus{
			Count:      deposits.Count,
			Processing: deposits.Active,
		},
		Batch: BatchStatus{
			Count:      batchs.Count,
			Processing: batchs.Active,
		},
		Withdraw: WithdrawStatus{
			Count:      witdthdraws.Count,
			Processing: witdthdraws.Active,
		},
	}, nil
}
