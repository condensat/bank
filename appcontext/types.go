package appcontext

import (
	"context"
)

type Worker interface {
	Run(ctx context.Context, numWorkers int)
}
