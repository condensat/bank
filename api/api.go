package api

import (
	"context"
)

type Api int

func (p *Api) Run(ctx context.Context) {

	<-ctx.Done()
}
