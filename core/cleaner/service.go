package cleaner

import (
	"context"
	"time"
)

type (
	Service struct {
		userStorage UserStorage
	}
)

func NewService(userStorage UserStorage) Service {
	return Service{
		userStorage: userStorage,
	}
}

const (
	activationExpireTime = 24 * time.Hour
)

func (s *Service) CleanExpiredUsers(ctx context.Context) error {
	if err := s.userStorage.DeleteUsersWithExpiredActivation(ctx, activationExpireTime); err != nil {
		return err
	}
	return nil
}
