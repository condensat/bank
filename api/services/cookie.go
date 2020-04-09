package services

import (
	"context"
	"net/http"

	"git.condensat.tech/bank/appcontext"
	"git.condensat.tech/bank/logger"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
)

type CookieCodec struct {
	ctx    context.Context
	codec  *json.Codec
	domain string
}

func NewCookieCodec(ctx context.Context) *CookieCodec {
	return &CookieCodec{
		ctx:    ctx,
		codec:  json.NewCodec(),
		domain: appcontext.Domain(ctx),
	}
}

func (p *CookieCodec) NewRequest(r *http.Request) rpc.CodecRequest {
	return &CookieCodecRequest{
		ctx:     p.ctx,
		request: p.codec.NewRequest(r),
		domain:  p.domain,
	}
}

type CookieCodecRequest struct {
	ctx     context.Context
	request rpc.CodecRequest
	domain  string
}

func (p *CookieCodecRequest) Method() (string, error) {
	return p.request.Method()
}

func (p *CookieCodecRequest) ReadRequest(args interface{}) error {
	return p.request.ReadRequest(args)
}

func (p *CookieCodecRequest) WriteResponse(w http.ResponseWriter, args interface{}) {
	if args == nil {
		return
	}

	switch reply := args.(type) {
	case *SessionReply:
		setSessionCookie(p.domain, w, reply)

	default:
		log := logger.Logger(p.ctx).WithField("Method", "CookieCodecRequest.WriteResponse")
		log.Debug("Unknwon Reply")
	}

	// forward to request
	p.request.WriteResponse(w, args)
}

func (p *CookieCodecRequest) WriteError(w http.ResponseWriter, status int, err error) {
	p.request.WriteError(w, status, err)
}
