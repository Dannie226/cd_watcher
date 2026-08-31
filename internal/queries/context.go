package queries

import (
	"context"
	"time"
)

func getTimeoutContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
	return ctx
}
