package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
	"x-bank-users/cerrors"
	"x-bank-users/core/web"
	"x-bank-users/entity"
	"x-bank-users/ercodes"
)

const (
	uniqueLoginConstraint = `users_login_key`
	uniqueEmailConstraint = `users_email_key`
)

type (
	Service struct {
		db *sql.DB
	}
)

func NewService(login, password, host string, port int, database string, maxCons int) (Service, error) {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", login, password, host, port, database))
	if err != nil {
		return Service{}, err
	}

	db.SetMaxOpenConns(maxCons)

	if err = db.Ping(); err != nil {
		return Service{}, err
	}

	return Service{
		db: db,
	}, err
}

func (s *Service) CreateUser(ctx context.Context, login, email string, passwordHash []byte) (int64, error) {
	const query = `INSERT INTO users (login, email, password) VALUES (@login, @email, @password) RETURNING id`

	row := s.db.QueryRowContext(ctx, query,
		pgx.NamedArgs{
			"login":    login,
			"email":    email,
			"password": passwordHash,
		},
	)

	if err := row.Err(); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.ConstraintName {
			case uniqueLoginConstraint:
				return 0, cerrors.NewErrorWithUserMessage(ercodes.LoginAlreadyTaken, nil, "Логин уже занят")
			case uniqueEmailConstraint:
				return 0, cerrors.NewErrorWithUserMessage(ercodes.EmailAlreadyTaken, nil, "Емейл уже занят")
			}
		}
		return 0, s.wrapQueryError(err)
	}

	var userId int64
	if err := row.Scan(&userId); err != nil {
		return 0, s.wrapScanError(err)
	}

	return userId, nil
}

func (s *Service) GetSignInDataByLogin(ctx context.Context, login string) (web.UserDataToSignIn, error) {
	var userData web.UserDataToSignIn

	const query = `SELECT users.id, users.password, users."telegramId", users_personal_data.id IS NOT NULL as "hasPersonalData"
				   FROM users
				   LEFT JOIN users_personal_data USING (id) 
				   WHERE users.login = @login`

	row := s.db.QueryRowContext(ctx, query,
		pgx.NamedArgs{
			"login": login,
		},
	)

	if err := row.Err(); err != nil {
		return web.UserDataToSignIn{}, s.wrapQueryError(err)
	}

	if err := row.Scan(&userData.Id, &userData.PasswordHash, &userData.TelegramId, &userData.HasPersonalData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return web.UserDataToSignIn{}, cerrors.NewErrorWithUserMessage(ercodes.InvalidLoginOrPassword, err, "Неверный логин пароль")
		}
		return web.UserDataToSignIn{}, s.wrapScanError(err)
	}

	return userData, nil
}

func (s *Service) GetSignInDataById(ctx context.Context, id int64) (web.UserDataToSignIn, error) {
	var userData web.UserDataToSignIn

	const query = `SELECT users.id, users.password, users."telegramId", users_personal_data.id IS NOT NULL as "hasUsersPersonalData" FROM users LEFT JOIN users_personal_data USING (id) WHERE id = @id`

	row := s.db.QueryRowContext(ctx, query,
		pgx.NamedArgs{
			"id": id,
		},
	)

	if err := row.Scan(&userData.Id, &userData.PasswordHash, &userData.TelegramId, &userData.HasPersonalData); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return web.UserDataToSignIn{}, s.wrapQueryError(err)
		}
		return web.UserDataToSignIn{}, s.wrapScanError(err)
	}

	return userData, nil
}

func (s *Service) UserIdByLoginAndEmail(ctx context.Context, login, email string) (int64, error) {
	const query = `SELECT id FROM users WHERE login = @login AND email = @email`

	row := s.db.QueryRowContext(ctx, query, pgx.NamedArgs{
		"login": login,
		"email": email,
	},
	)

	err := row.Err()
	if err != nil {
		return 0, s.wrapQueryError(err)

	}

	var userId int64
	err = row.Scan(&userId)
	if err != nil {
		return 0, s.wrapScanError(err)
	}
	return userId, nil
}

func (s *Service) UpdatePassword(ctx context.Context, id int64, passwordHash []byte) error {
	const query = `UPDATE users SET password = @password WHERE id = @id`

	_, err := s.db.ExecContext(ctx, query, pgx.NamedArgs{
		"id":       id,
		"password": passwordHash,
	},
	)

	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) UpdateTelegramId(ctx context.Context, telegramId *int64, userId int64) error {
	const query = `UPDATE users SET "telegramId" = @telegramId WHERE id = @id`

	_, err := s.db.ExecContext(ctx, query, pgx.NamedArgs{
		"id":         userId,
		"telegramId": telegramId,
	},
	)

	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) GetUserPersonalDataById(ctx context.Context, userId int64) (*web.UserPersonalData, error) {
	const query = `SELECT "phoneNumber", "firstName", "lastName", "fathersName", "dateOfBirth", "passportId", "address", gender, countries.name FROM users_personal_data JOIN countries on users_personal_data."liveInCountry" = countries.id where users_personal_data."id" = $1`

	row := s.db.QueryRowContext(ctx, query, userId)

	if err := row.Err(); err != nil {
		return nil, s.wrapQueryError(err)
	}
	row = s.db.QueryRowContext(ctx, query, userId)

	var userPersonalData web.UserPersonalData
	err := row.Scan(&userPersonalData.PhoneNumber, &userPersonalData.FirstName, &userPersonalData.LastName, &userPersonalData.FathersName, &userPersonalData.DateOfBirth, &userPersonalData.PassportId, &userPersonalData.Address, &userPersonalData.Gender, &userPersonalData.LiveInCountry)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, s.wrapScanError(err)
	}

	return &userPersonalData, nil
}

func (s *Service) DeleteUsersWithExpiredActivation(ctx context.Context, expirationTime time.Duration) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE "createdAt" < $1`, time.Now().Add(-expirationTime))

	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) GetUserDataById(ctx context.Context, id int64) (web.UserData, error) {
	const query = `SELECT id, uuid, login, email, "telegramId", "createdAt" FROM users WHERE id = @id`

	row := s.db.QueryRowContext(ctx, query, pgx.NamedArgs{
		"id": id,
	},
	)

	if err := row.Err(); err != nil {
		return web.UserData{}, s.wrapQueryError(err)
	}

	var userData web.UserData
	err := row.Scan(&userData.Id, &userData.UUID, &userData.Login, &userData.Email, &userData.TelegramId, &userData.CreatedAt)
	if err != nil {
		return web.UserData{}, s.wrapScanError(err)
	}

	return userData, nil
}

func (s *Service) AddUsersAuthHistory(ctx context.Context, userId int64, agent, ip string) error {
	const query = `INSERT INTO users_auth_history ("userId", "agent", ip) VALUES (@userId, @agent, @ip)`

	_, err := s.db.ExecContext(ctx, query,
		pgx.NamedArgs{
			"userId": userId,
			"agent":  agent,
			"ip":     ip,
		},
	)
	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) GetUserAuthHistory(ctx context.Context, userId int64) ([]web.UserAuthHistoryData, error) {
	const query = `SELECT "userId", "agent", "ip", "timestamp" FROM users_auth_history WHERE "userId" = $1 ORDER BY timestamp DESC `

	rows, err := s.db.QueryContext(ctx, query, userId)

	if err != nil {
		return nil, s.wrapQueryError(err)
	}

	var userAuthHistoryData []web.UserAuthHistoryData
	for rows.Next() {
		var userAuthHist web.UserAuthHistoryData
		if err = rows.Scan(&userAuthHist.Id, &userAuthHist.Agent, &userAuthHist.Ip, &userAuthHist.Timestamp); err != nil {
			return nil, s.wrapScanError(err)
		}
		userAuthHistoryData = append(userAuthHistoryData, userAuthHist)
	}

	return userAuthHistoryData, nil
}

func (s *Service) AddUserPersonalDataById(_ context.Context, userId int64, data entity.UserPersonalData) error {
	const query = `
INSERT INTO users_personal_data 
    (id, "phoneNumber", "firstName", "lastName", "fathersName", "dateOfBirth", "passportId", address, gender, "liveInCountry") 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := s.db.Exec(query, userId, data.PhoneNumber, data.FirstName, data.LastName, data.FathersName, data.DateOfBirth,
		data.PassportId, data.Address, data.Gender, data.LiveInCountryId)

	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}

func (s *Service) UpdateUserPersonalDataById(_ context.Context, userId int64, data entity.UserPersonalData) error {
	const query = `
UPDATE users_personal_data 
SET "phoneNumber" = $1, 
    "firstName" = $2, 
    "lastName" = $3, 
    "fathersName" = $4, 
    "dateOfBirth" = $5, 
    "passportId" = $6, 
    address = $7, 
    gender = $8, 
    "liveInCountry" = $9
WHERE id = $10`

	_, err := s.db.Exec(query, data.PhoneNumber, data.FirstName, data.LastName, data.FathersName, data.DateOfBirth,
		data.PassportId, data.Address, data.Gender, data.LiveInCountryId, userId)

	if err != nil {
		return s.wrapQueryError(err)
	}

	return nil
}
