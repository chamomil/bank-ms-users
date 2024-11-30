package redis

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"strings"
	"time"
	"x-bank-users/cerrors"
	"x-bank-users/ercodes"
)

type (
	Service struct {
		db *redis.Client
	}
)

const (
	refreshTokenScanSize = 10
)

func NewService(password, host string, port int, database, maxCons int) (Service, error) {
	client := redis.NewClient(&redis.Options{
		Addr:           host + ":" + strconv.Itoa(port),
		Password:       password,
		DB:             database,
		MaxActiveConns: maxCons,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return Service{}, err
	}

	return Service{
		db: client,
	}, nil
}

func (s *Service) SaveActivationCode(ctx context.Context, code string, userId int64, ttl time.Duration) error {
	if err := s.db.Set(ctx, activationCodeKey+code, userId, ttl).Err(); err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) VerifyActivationCode(ctx context.Context, code string) (int64, error) {
	userId, err := s.db.Get(ctx, activationCodeKey+code).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cerrors.NewErrorWithUserMessage(ercodes.ActivationCodeNotFound, nil, "Код активации не найден")
		}
		return 0, s.wrapQueryError(err)
	}
	return userId, nil
}

func (s *Service) SaveRecoveryCode(ctx context.Context, code string, userId int64, ttl time.Duration) error {
	if err := s.db.Set(ctx, recoveryCodeKey+code, userId, ttl).Err(); err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) VerifyRecoveryCode(ctx context.Context, code string) (int64, error) {
	userId, err := s.db.Get(ctx, recoveryCodeKey+code).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cerrors.NewErrorWithUserMessage(ercodes.RecoveryCodeNotFound, nil, "Код восстановления не найден")
		}
		return 0, s.wrapQueryError(err)
	}

	return userId, nil
}

func (s *Service) SaveRefreshToken(ctx context.Context, token string, userId int64, ttl time.Duration) error {
	if err := s.db.Set(ctx, refreshTokenKey+token, userId, ttl).Err(); err != nil {
		return s.wrapQueryError(err)
	}
	if err := s.db.Set(ctx, userRefreshTokenKey+strconv.FormatInt(userId, 10)+":"+token, true, ttl).Err(); err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) VerifyRefreshToken(ctx context.Context, token string) (int64, error) {
	userId, err := s.db.Get(ctx, refreshTokenKey+token).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cerrors.NewErrorWithUserMessage(ercodes.RefreshTokenNotFound, nil, "Токен не найден")
		}
		return 0, s.wrapQueryError(err)
	}

	return userId, nil
}

func (s *Service) ExpireAllByUserId(ctx context.Context, userId int64) error {
	var cursor uint64
	var keys []string
	var err error
	var keysToDelete []string

	for {
		keys, cursor, err = s.db.Scan(ctx, cursor, userRefreshTokenKey+strconv.FormatInt(userId, 10)+":*", refreshTokenScanSize).Result()
		if err != nil {
			return cerrors.NewErrorWithUserMessage(ercodes.ExpireAllByUserIdError, err, "Ошибка сканирования токенов")
		}
		for _, key := range keys {
			token := strings.Split(key, ":")[3]
			keysToDelete = append(keysToDelete, refreshTokenKey+token, key)
		}
		if cursor == 0 {
			break
		}
	}

	if keysToDelete == nil {
		return nil
	}

	if err := s.db.Del(ctx, keysToDelete...).Err(); err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.ExpireAllByUserIdError, err, "Ошибка удаления токенов")
	}
	return nil
}

func (s *Service) Save2FaCode(ctx context.Context, code string, userId int64, ttl time.Duration) error {
	if err := s.db.Set(ctx, TwoFaCodeKey+code, userId, ttl).Err(); err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) Verify2FaCode(ctx context.Context, code string) (int64, error) {
	userId, err := s.db.Get(ctx, TwoFaCodeKey+code).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, cerrors.NewErrorWithUserMessage(ercodes.TwoFaCodeNotFound, nil, "Код двухфакторной аутентификации не найден")
		}
		return 0, s.wrapQueryError(err)
	}

	return userId, nil
}
