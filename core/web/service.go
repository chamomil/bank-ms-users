package web

import (
	"context"
	"github.com/google/uuid"
	"time"
	"x-bank-users/auth"
	"x-bank-users/cerrors"
	"x-bank-users/entity"
	"x-bank-users/ercodes"
)

type (
	Service struct {
		userStorage           UserStorage
		randomGenerator       RandomGenerator
		activationCodeCache   ActivationCodeStorage
		passwordHasher        PasswordHasher
		refreshTokenStorage   RefreshTokenStorage
		twoFactorCodeStorage  TwoFactorCodeStorage
		twoFactorCodeNotifier TwoFactorCodeNotifier
		recoveryCodeStorage   RecoveryCodeStorage
	}
)

func NewService(
	userStorage UserStorage,
	randomGenerator RandomGenerator,
	activationCodeCache ActivationCodeStorage,
	passwordHasher PasswordHasher,
	refreshTokenStorage RefreshTokenStorage,
	twoFactorCodeStorage TwoFactorCodeStorage,
	twoFactorCodeNotifier TwoFactorCodeNotifier,
	recoveryCodeStorage RecoveryCodeStorage,
) Service {
	return Service{
		userStorage:           userStorage,
		randomGenerator:       randomGenerator,
		activationCodeCache:   activationCodeCache,
		passwordHasher:        passwordHasher,
		refreshTokenStorage:   refreshTokenStorage,
		twoFactorCodeStorage:  twoFactorCodeStorage,
		twoFactorCodeNotifier: twoFactorCodeNotifier,
		recoveryCodeStorage:   recoveryCodeStorage,
	}
}

const (
	hashCost = 10

	claimsTtl = time.Minute * 5

	refreshTokenCharset = ".-"
	refreshTokenSize    = 2048
	refreshTokenTtl     = time.Hour * 24 * 7

	twoFactorCodeCharset = "0123456789"
	twoFactorCodeSize    = 6
	TwoFactorCodeTtl     = time.Minute * 5

	recoveryCodeCharset = "ij"
	recoveryCodeSize    = 16
	recoveryCodeTtl     = time.Minute * 5
)

func (s *Service) SignUp(ctx context.Context, login, password, email string) error {
	hash, err := s.passwordHasher.HashPassword(ctx, []byte(password), hashCost)
	if err != nil {
		return err
	}

	_, err = s.userStorage.CreateUser(ctx, login, email, hash)
	if err != nil {
		return err
	}
	return err
}

func (s *Service) SignIn(ctx context.Context, login, password, agent, ip string) (SignInResult, error) {
	userData, err := s.userStorage.GetSignInDataByLogin(ctx, login)
	if err != nil {
		return SignInResult{}, err
	}

	err = s.passwordHasher.CompareHashAndPassword(ctx, password, userData.PasswordHash)
	if err != nil {
		return SignInResult{}, err
	}

	var refreshToken string
	refreshToken, err = s.getNewToken(ctx, userData.Id)
	if err != nil {
		return SignInResult{}, err
	}

	if err = s.userStorage.AddUsersAuthHistory(ctx, userData.Id, agent, ip); err != nil {
		return SignInResult{}, err
	}
	date := time.Now()

	claims := auth.Claims{
		Id:              uuid.New().String(),
		IssuedAt:        date.Unix(),
		ExpiresAt:       date.Add(claimsTtl).Unix(),
		Sub:             userData.Id,
		Is2FAToken:      false,
		HasPersonalData: userData.HasPersonalData,
	}

	return SignInResult{
		AccessClaims: claims,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) SignIn2FA(ctx context.Context, claims auth.Claims, code, agent, ip string) (SignInResult, error) {
	userId, err := s.twoFactorCodeStorage.Verify2FaCode(ctx, code)
	if err != nil {
		return SignInResult{}, err
	}

	if userId != claims.Sub {
		return SignInResult{}, cerrors.NewErrorWithUserMessage(ercodes.Invalid2FACode, nil, "Неверный 2FA код")
	}

	personalData, err := s.userStorage.GetSignInDataById(ctx, userId)
	if err != nil {
		return SignInResult{}, err
	}

	if err = s.userStorage.AddUsersAuthHistory(ctx, personalData.Id, agent, ip); err != nil {
		return SignInResult{}, err
	}

	hasPersonalData := personalData.HasPersonalData

	refreshToken, err := s.randomGenerator.GenerateString(ctx, refreshTokenCharset, refreshTokenSize)
	if err != nil {
		return SignInResult{}, err
	}

	err = s.refreshTokenStorage.SaveRefreshToken(ctx, refreshToken, userId, refreshTokenTtl)
	if err != nil {
		return SignInResult{}, err
	}

	timeNow := time.Now()
	accessClaims := auth.Claims{
		Id:              uuid.New().String(),
		IssuedAt:        timeNow.Unix(),
		ExpiresAt:       timeNow.Add(claimsTtl).Unix(),
		Sub:             userId,
		Is2FAToken:      false,
		HasPersonalData: hasPersonalData,
	}

	return SignInResult{
		AccessClaims: accessClaims,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Recovery(ctx context.Context, login, email string) error {
	userId, err := s.userStorage.UserIdByLoginAndEmail(ctx, login, email)
	if err != nil {
		return err
	}

	recoveryCode, err := s.randomGenerator.GenerateString(ctx, recoveryCodeCharset, recoveryCodeSize)
	if err != nil {
		return err
	}

	err = s.recoveryCodeStorage.SaveRecoveryCode(ctx, recoveryCode, userId, recoveryCodeTtl)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) RecoveryCode(ctx context.Context, code, password string) error {
	userId, err := s.recoveryCodeStorage.VerifyRecoveryCode(ctx, code)
	if err != nil {
		return err
	}

	hashedPassword, err := s.passwordHasher.HashPassword(ctx, []byte(password), hashCost)
	if err != nil {
		return err
	}

	err = s.userStorage.UpdatePassword(ctx, userId, hashedPassword)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Refresh(ctx context.Context, token string) (SignInResult, error) {
	userId, err := s.refreshTokenStorage.VerifyRefreshToken(ctx, token)
	if err != nil {
		return SignInResult{}, err
	}

	refreshToken, err := s.getNewToken(ctx, userId)
	if err != nil {
		return SignInResult{}, err
	}
	userData, err := s.userStorage.GetSignInDataById(ctx, userId)
	if err != nil {
		return SignInResult{}, err
	}

	date := time.Now()

	claims := auth.Claims{
		Id:              uuid.New().String(),
		IssuedAt:        date.Unix(),
		ExpiresAt:       date.Add(claimsTtl).Unix(),
		Sub:             userId,
		Is2FAToken:      false,
		HasPersonalData: userData.HasPersonalData,
	}

	return SignInResult{
		AccessClaims: claims,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) getNewToken(ctx context.Context, userId int64) (string, error) {
	err := s.refreshTokenStorage.ExpireAllByUserId(ctx, userId)
	if err != nil {
		return "", err
	}
	refreshToken, err := s.randomGenerator.GenerateString(ctx, refreshTokenCharset, refreshTokenSize)
	if err != nil {
		return "", err
	}
	if err = s.refreshTokenStorage.SaveRefreshToken(ctx, refreshToken, userId, refreshTokenTtl); err != nil {
		return "", err
	}
	return refreshToken, nil
}

func (s *Service) BindTelegram(ctx context.Context, telegramId *int64, userId int64) error {
	return s.userStorage.UpdateTelegramId(ctx, telegramId, userId)
}

func (s *Service) DeleteTelegram(ctx context.Context, userId int64) error {
	return s.userStorage.UpdateTelegramId(ctx, nil, userId)
}

func (s *Service) GetUserPersonalData(ctx context.Context, userId int64) (*UserPersonalData, error) {
	return s.userStorage.GetUserPersonalDataById(ctx, userId)
}

func (s *Service) AddUserPersonalData(ctx context.Context, userId int64, data entity.UserPersonalData) error {
	exist, err := s.GetUserPersonalData(ctx, userId)
	if err != nil {
		return err
	}
	if exist != nil {
		return s.userStorage.UpdateUserPersonalDataById(ctx, userId, data)
	}
	return s.userStorage.AddUserPersonalDataById(ctx, userId, data)
}

func (s *Service) GetUserData(ctx context.Context, userId int64) (UserData, error) {
	return s.userStorage.GetUserDataById(ctx, userId)
}

func (s *Service) GetAuthHistory(ctx context.Context, userId int64) ([]UserAuthHistoryData, error) {
	return s.userStorage.GetUserAuthHistory(ctx, userId)
}

func (s *Service) GetWorkplaces(ctx context.Context, userId int64) ([]entity.UserWorkplace, error) {
	return s.userStorage.GetUserWorkplaces(ctx, userId)
}

func (s *Service) AddWorkplace(ctx context.Context, userId int64, work entity.Workplace) error {
	return s.userStorage.AddUserWorkplace(ctx, userId, work)
}

func (s *Service) GetCountries(ctx context.Context) ([]entity.Country, error) {
	return s.userStorage.GetCountries(ctx)
}
