package redis

import (
	"x-bank-users/cerrors"
	"x-bank-users/ercodes"
)

func (s *Service) Close() {
	_ = s.db.Close()
}

func (s *Service) wrapQueryError(err error) error {
	return cerrors.NewErrorWithUserMessage(ercodes.RedisQuery, err, "Ошибка работы с базой данных")
}
