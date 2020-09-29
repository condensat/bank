package services

import (
	"context"
	"sort"

	"git.condensat.tech/bank/utils"

	wallet "git.condensat.tech/bank/wallet/client"
)

type WalletInfo struct {
	UTXOs  int     `json:"utxos"`
	Amount float64 `json:"amount"`
}

type WalletStatus struct {
	Chain  string     `json:"chain"`
	Asset  string     `json:"asset"`
	Total  WalletInfo `json:"total"`
	Locked WalletInfo `json:"locked"`
}

type ReserveStatus struct {
	Wallets []WalletStatus `json:"wallets"`
}

func FetchReserveStatus(ctx context.Context) (ReserveStatus, error) {
	walletStatus, err := wallet.WalletStatus(ctx)
	if err != nil {
		return ReserveStatus{}, err
	}

	var wallets []WalletStatus
	assetMap := make(map[string]*WalletStatus)
	for _, wallet := range walletStatus.Wallets {
		for _, utxo := range wallet.UTXOs {

			// get or create WalletStatus from assetMap
			key := wallet.Chain + utxo.Asset
			ws, ok := assetMap[key]
			if !ok {
				ws = &WalletStatus{
					Chain: wallet.Chain,
					Asset: utxo.Asset,
				}
				assetMap[key] = ws
			}

			ws.Total.Amount += utxo.Amount
			ws.Total.UTXOs++
			if utxo.Locked {
				ws.Locked.Amount += utxo.Amount
				ws.Locked.UTXOs++
			}
		}
	}

	for _, ws := range assetMap {
		ws.Total.Amount = utils.ToFixed(ws.Total.Amount, 8)
		ws.Locked.Amount = utils.ToFixed(ws.Locked.Amount, 8)

		wallets = append(wallets, *ws)
	}

	// Sort wallets
	sort.Slice(wallets, func(i, j int) bool {
		if wallets[i].Chain != wallets[j].Chain {
			return wallets[i].Chain < wallets[j].Chain
		}

		return wallets[i].Asset < wallets[j].Asset
	})

	return ReserveStatus{
		Wallets: wallets,
	}, nil
}
