package handlers

import (
	"context"
	"errors"
	"time"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/logger"

	"git.condensat.tech/bank/swap/liquid/common"

	"git.condensat.tech/bank/cache"
	"git.condensat.tech/bank/messaging"

	"git.condensat.tech/bank/utils/shellexec"

	"github.com/sirupsen/logrus"
)

func FinalizeSwapProposal(ctx context.Context, swapID uint64, payload common.Payload) (common.SwapProposal, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.handler.FinalizeSwapProposal")

	log = log.WithField("SwapID", swapID)

	if !payload.Valid() {
		log.WithError(common.ErrInvalidPayload).
			WithField("Payload", payload).
			Error("Invalid Payload")
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	result := common.SwapProposal{
		Timestamp: time.Now().UTC().Truncate(time.Millisecond),
		SwapID:    swapID,
	}

	ShellExecLock.Lock()
	defer ShellExecLock.Unlock()

	out, err := shellexec.Execute(ctx, LiquidSwapFinalize(payload))
	if len(out.Stdout) == 0 && err == nil {
		err = errors.New("No Output")
	}
	if err != nil {
		log.WithError(err).
			WithFields(logrus.Fields{
				"Stdout": out.Stdout,
				"Stderr": out.Stderr,
				"Code":   out.Code,
			}).
			Error("out")
		return result, err
	}

	result.Payload = common.Payload(out.Stdout)

	if !result.Payload.Valid() {
		log.WithError(common.ErrInvalidPayload).
			WithField("Payload", result.Payload).
			Error("Invalid Payload")
		return common.SwapProposal{}, common.ErrInvalidPayload
	}

	log.WithField("Result", result).
		Debug("Finalize Swap Proposal")

	return result, nil
}

func OnFinalizeSwapProposal(ctx context.Context, subject string, message *bank.Message) (*bank.Message, error) {
	log := logger.Logger(ctx).WithField("Method", "Liquid.handler.OnFinalizeSwapProposal")
	log = log.WithFields(logrus.Fields{
		"Subject": subject,
	})

	var request common.SwapProposal
	return messaging.HandleRequest(ctx, message, &request,
		func(ctx context.Context, _ bank.BankObject) (bank.BankObject, error) {
			log = log.WithFields(logrus.Fields{
				"SwapID": request.SwapID,
			})

			response, err := FinalizeSwapProposal(ctx, request.SwapID, request.Payload)
			if err != nil {
				log.WithError(err).
					Errorf("Failed to FinalizeSwapProposal")
				return nil, cache.ErrInternalError
			}

			// create & return response
			return &response, nil
		})
}
