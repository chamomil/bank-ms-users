package web

import (
	"time"
	"x-bank-users/auth"
)

type (
	UserDataToSignIn struct {
		Id              int64
		PasswordHash    []byte
		TelegramId      *int64
		IsActivated     bool
		HasPersonalData bool
	}

	SignInResult struct {
		AccessClaims auth.Claims
		RefreshToken string
	}

	UserPersonalData struct {
		Id              int64
		PhoneNumber     string
		FirstName       string
		LastName        string
		FathersName     *string
		DateOfBirth     time.Time
		PassportId      string
		Address         string
		Gender          string
		LiveInCountry   string
		UserEmployments []UserEmployment
	}

	UserEmployment struct {
		Workplace Workplace
		Position  string
		StartDate time.Time
		EndDate   time.Time
	}

	Workplace struct {
		Name    string
		Address string
	}

	UserData struct {
		Id         int64
		UUID       string
		Login      string
		Email      string
		TelegramId *int64
		CreatedAt  time.Time
	}

	UserAuthHistoryData struct {
		Id        int64
		Agent     string
		Ip        string
		Timestamp time.Time
	}
)
