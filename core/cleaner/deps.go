package cleaner

import (
	"context"
	"time"
)

type (
	UserStorage interface {
		DeleteUsersWithExpiredActivation(ctx context.Context, expirationTime time.Duration) error
	}
)
