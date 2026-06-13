package postgres

import (
	"context"
	"time"
)

const queryTimeout = 3 * time.Second

func newQueryContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), queryTimeout)
}
