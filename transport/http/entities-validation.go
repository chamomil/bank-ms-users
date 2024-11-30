package http

import (
	"net/http"
	"regexp"
)

var (
	isValidEmail = regexp.MustCompile("^.+@.+\\..+$").MatchString
	isValidLogin = regexp.MustCompile("^[a-z0-9_-]{6,32}$").MatchString
)

func (t *Transport) validate(w http.ResponseWriter, v validatable) bool {
	ve := v.validate()
	if len(ve) > 0 {
		t.errorHandler.setUnprocessableEntityError(w, ve)
		return false
	}

	return true
}

func (u *UserDataToSignUp) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 3)

	if !isValidEmail(u.Email) {
		ve.Add("Неверный адрес электронной почты")
	}

	if !isValidLogin(u.Login) {
		ve.Add("Неверный логин")
	}

	if len(u.Password) < 6 {
		ve.Add("Слишком короткий пароль")
	} else if len(u.Password) > 16 {
		ve.Add("Слишком длинный пароль")
	}

	return
}

func (u *UserDataToSignIn) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 2)

	if !isValidLogin(u.Login) {
		ve.Add("Неверный логин")
	}

	if len(u.Password) < 6 || len(u.Password) > 16 {
		ve.Add("Неверный пароль")
	}

	return
}

func (u *RecoveryRequest) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 2)

	if !isValidLogin(u.Login) {
		ve.Add("Неверный логин")
	}

	if !isValidEmail(u.Email) {
		ve.Add("Неверный адрес электронной почты")
	}

	return
}

func (u *RecoveryCodeRequest) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 1)

	if len(u.Password) < 6 || len(u.Password) > 16 {
		ve.Add("Неверный пароль")
	}

	return
}

func (u *TelegramBindRequest) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 3)

	if u.TelegramId == 0 {
		ve.Add("Неверный id")
	}
	if len(u.FirstName) == 0 {
		ve.Add("Неверное имя пользователя")
	}
	if len(u.LastName) == 0 {
		ve.Add("Неверная фамилия пользователя")
	}
	if len(u.Username) == 0 {
		ve.Add("Неверный username")
	}
	if len(u.PhotoUrl) == 0 {
		ve.Add("Неверный путь к фотографии")
	}
	if u.AuthDate == 0 {
		ve.Add("Неверная дата авторизаци")
	}
	if len(u.Hash) == 0 {
		ve.Add("Неверный хэш")
	}

	return
}
