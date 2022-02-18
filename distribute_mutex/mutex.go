package distribute_mutex

import (
	"context"
)

type IMutex interface {
	// block
	Lock(ctx context.Context) error
	// non-block
	TryLock(ctx context.Context) error
	UnLock(ctx context.Context) error
	// mutex kind
	Kind() string
}
