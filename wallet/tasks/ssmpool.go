package tasks

import (
	"context"
	"fmt"
	"time"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/wallet/common"
	"git.condensat.tech/bank/wallet/ssm/commands"

	"git.condensat.tech/bank/database"
	"git.condensat.tech/bank/database/model"

	"github.com/sirupsen/logrus"
)

type SsmInfo struct {
	Device      string
	Chain       string
	Fingerprint string
}

// SsmPool
func SsmPool(ctx context.Context, epoch time.Time, infos []SsmInfo) {
	log := logger.Logger(ctx).WithField("Method", "task.SsmPool")
	db := appcontext.Database(ctx)

	log.WithFields(logrus.Fields{
		"Epoch": epoch.Truncate(time.Millisecond),
		"Count": len(infos),
	}).Info("Maintain ssm pool addresses")

	for _, info := range infos {
		ssm := common.SsmClientFromContext(ctx, info.Device)

		// Todo: count unused ssm addresses from database

		// count actual ssm addresses count for chain/fingerprint
		addressCount, err := database.CountSsmAddress(db,
			model.SsmChain(info.Chain),
			model.SsmFingerprint(info.Fingerprint),
		)
		if err != nil {
			log.WithError(err).Error("CountSsmAddress failed")
			continue
		}

		// create new address for next path
		// Todo: manage annual rotation for path
		nextPath := fmt.Sprintf("84h/0h/%d", addressCount+1)

		ssmAddress, err := ssm.NewAddress(ctx, commands.SsmPath{
			Chain:       info.Chain,
			Fingerprint: info.Fingerprint,
			Path:        nextPath,
		})
		if err != nil {
			log.WithError(err).Error("NewAddress failed")
			continue
		}
		if info.Chain != ssmAddress.Address {
			if err != nil {
				log.WithError(err).Error("Wrong ssmAddress chain")
				continue
			}
		}

		// store new address to database
		ssmAddressID, err := database.AddSsmAddress(db,
			model.SsmAddress{
				PublicAddress: model.SsmPublicAddress(ssmAddress.Address),
				ScriptPubkey:  model.SsmPubkey(ssmAddress.PubKey),
				BlindingKey:   model.SsmBlindingKey(ssmAddress.BlindingKey),
			},
			model.SsmAddressInfo{
				Chain:       model.SsmChain(info.Chain),
				Fingerprint: model.SsmFingerprint(info.Fingerprint),
				HDPath:      model.SsmHDPath(nextPath),
			},
		)
		if err != nil {
			log.WithError(err).Error("AddSsmAddress failed")
			continue
		}

		log.
			WithField("ssmAddressID", ssmAddressID).
			Debug("New ssm address")
	}
}
