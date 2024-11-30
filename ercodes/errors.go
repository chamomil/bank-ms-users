package ercodes

import "x-bank-users/cerrors"

const (
	_ cerrors.Code = -iota

	UserNotFound
	ActivationCodeNotFound
	BcryptHashing
	RandomGeneration
	HS512Authorization
	RS256Authorization
	AccountNotActivated
	WrongPassword
	Invalid2FACode
	GmailSendError
	LoginAlreadyTaken
	EmailAlreadyTaken
	PostgresQuery
	PostgresScan
	RedisQuery
	RecoveryCodeNotFound
	RefreshTokenNotFound
	TwoFaCodeNotFound
	ExpireAllByUserIdError
	InvalidLoginOrPassword
	TelegramSendError
)
