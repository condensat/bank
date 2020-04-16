package messaging

import (
	"context"
	"errors"

	"git.condensat.tech/bank"
	"git.condensat.tech/bank/appcontext"
)

var (
	ErrHandleRequest = errors.New("Request handle failed")
	ErrRequestFailed = errors.New("Request Failed")
)

type RequestHandler func(ctx context.Context, request bank.BankObject) (bank.BankObject, error)

func HandleRequest(ctx context.Context, message *bank.Message, request bank.BankObject, handle RequestHandler) (*bank.Message, error) {
	err := bank.FromMessage(message, request)
	if err != nil {
		return nil, err
	}

	resp, err := handle(ctx, request)
	if err != nil {
		return nil, err
	}

	message = bank.ToMessage(appcontext.AppName(ctx), resp)
	if message == nil {
		err = ErrHandleRequest
	}

	return message, err
}

func RequestMessage(ctx context.Context, subject string, req, resp bank.BankObject) error {
	messaging := appcontext.Messaging(ctx)

	message := bank.ToMessage(appcontext.AppName(ctx), req)

	message, err := messaging.Request(ctx, subject, message)
	if err != nil {
		return err
	}

	err = bank.FromMessage(message, resp)
	if err != nil {
		return err
	}

	return nil
}
