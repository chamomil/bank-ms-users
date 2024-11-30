package http

type (
	UserDataToSignUp struct {
		Email    string `json:"email"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	UserDataToSignIn struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	SignInResponse struct {
		TwoFaDemand string    `json:"2FA"`
		Tokens      TokenPair `json:"tokens"`
	}

	TokenPair struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	UserDataToSignIn2FA struct {
		Code string `json:"code"`
	}

	RecoveryRequest struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}

	RecoveryCodeRequest struct {
		Password string `json:"password"`
	}

	RefreshRequest struct {
		RefreshToken string `json:"refreshToken"`
	}

	TelegramBindRequest struct {
		TelegramId int64  `json:"id"`
		FirstName  string `json:"firstname"`
		LastName   string `json:"lastname"`
		Username   string `json:"username"`
		PhotoUrl   string `json:"photoUrl"`
		AuthDate   int64  `json:"authDate"`
		Hash       string `json:"hash"`
	}

	UserPersonalData struct {
		PhoneNumber   string  `json:"phoneNumber"`
		FirstName     string  `json:"firstName"`
		LastName      string  `json:"lastName"`
		FathersName   *string `json:"fathersName"`
		DateOfBirth   string  `json:"dateOfBirth"`
		PassportId    string  `json:"passportId"`
		Address       string  `json:"address"`
		Gender        string  `json:"gender"`
		LiveInCountry string  `json:"liveInCountry"`
	}

	UserPersonalDataResponse struct {
		PersonalData *UserPersonalData `json:"personalData"`
	}

	UserDataResponse struct {
		Id         int64  `json:"id"`
		UUID       string `json:"uuid"`
		Login      string `json:"login"`
		Email      string `json:"email"`
		TelegramId *int64 `json:"telegramId"`
		CreatedAt  string `json:"createdAt"`
	}

	UserAuthHistoryResponseItem struct {
		Id        int64  `json:"id"`
		Agent     string `json:"agent"`
		Ip        string `json:"ip"`
		Timestamp string `json:"timestamp"`
	}

	UserAuthHistoryResponse struct {
		Items []UserAuthHistoryResponseItem `json:"items"`
	}
)
