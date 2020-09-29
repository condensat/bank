package services

import (
	"context"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/database"
)

type SwapStatus struct {
	Count      int `json:"count"`
	Processing int `json:"processing"`
}

func FetchSwapStatus(ctx context.Context) (SwapStatus, error) {
	db := appcontext.Database(ctx)

	swaps, err := database.SwapssInfos(db)
	if err != nil {
		return SwapStatus{}, err
	}

	return SwapStatus{
		Count:      swaps.Count,
		Processing: swaps.Active,
	}, nil
}
