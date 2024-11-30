package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"x-bank-users/auth"
)

func (t *Transport) handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	t.errorHandler.setNotFoundError(w)
}

func (t *Transport) handlerSignUp(w http.ResponseWriter, r *http.Request) {
	userData := UserDataToSignUp{}

	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &userData) {
		return
	}

	if err := t.service.SignUp(r.Context(), userData.Login, userData.Password, userData.Email); err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (t *Transport) handlerActivateAccount(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	err := t.service.ActivateAccount(r.Context(), code)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerSignIn(w http.ResponseWriter, r *http.Request) {
	userDataToSignIn := UserDataToSignIn{}
	if err := json.NewDecoder(r.Body).Decode(&userDataToSignIn); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &userDataToSignIn) {
		return
	}

	agent := r.Header.Get("User-Agent")
	ip := r.Header.Get("X-Real-Ip")

	signInResult, err := t.service.SignIn(r.Context(), userDataToSignIn.Login, userDataToSignIn.Password, agent, ip)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	token, err := t.authorizer.Authorize(r.Context(), signInResult.AccessClaims)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
	signInResponse := SignInResponse{}

	if signInResult.AccessClaims.Is2FAToken {
		signInResponse.TwoFaDemand = string(token)
	} else {
		signInResponse.Tokens.AccessToken = string(token)
		signInResponse.Tokens.RefreshToken = signInResult.RefreshToken
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(signInResponse)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerSignIn2FA(w http.ResponseWriter, r *http.Request) {
	userDataToSignIn2FA := UserDataToSignIn2FA{}

	err := json.NewDecoder(r.Body).Decode(&userDataToSignIn2FA)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	code := userDataToSignIn2FA.Code
	agent := r.Header.Get("User-Agent")
	ip := r.Header.Get("X-Real-Ip")

	signInResult, err := t.service.SignIn2FA(r.Context(), *claims, code, agent, ip)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	token, err := t.authorizer.Authorize(r.Context(), signInResult.AccessClaims)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	signInResponse := SignInResponse{}

	signInResponse.Tokens.AccessToken = string(token)
	signInResponse.Tokens.RefreshToken = signInResult.RefreshToken

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(signInResponse)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerRecovery(w http.ResponseWriter, r *http.Request) {
	var request RecoveryRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &request) {
		return
	}

	err = t.service.Recovery(r.Context(), request.Login, request.Email)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerRecoveryCode(w http.ResponseWriter, r *http.Request) {
	var request RecoveryCodeRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &request) {
		return
	}

	code := r.PathValue("code")

	err = t.service.RecoveryCode(r.Context(), code, request.Password)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	var request RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	signInResult, err := t.service.Refresh(r.Context(), request.RefreshToken)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
	token, err := t.authorizer.Authorize(r.Context(), signInResult.AccessClaims)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	refreshResponse := SignInResponse{
		Tokens: TokenPair{
			RefreshToken: signInResult.RefreshToken,
			AccessToken:  string(token),
		},
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(refreshResponse)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerGetUserPersonalData(w http.ResponseWriter, r *http.Request) {
	var userData UserPersonalData
	var response UserPersonalDataResponse
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	userId := claims.Sub

	data, err := t.service.GetUserPersonalData(r.Context(), userId)

	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	if data != nil {
		userData = UserPersonalData{
			PhoneNumber:   data.PhoneNumber,
			FirstName:     data.FirstName,
			LastName:      data.LastName,
			FathersName:   data.FathersName,
			DateOfBirth:   data.DateOfBirth.Format("2006-01-02"),
			PassportId:    data.PassportId,
			Address:       data.Address,
			Gender:        data.Gender,
			LiveInCountry: data.LiveInCountry,
		}

		response.PersonalData = &userData
	} else {
		response.PersonalData = nil
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerGetUserData(w http.ResponseWriter, r *http.Request) {
	var userData UserDataResponse

	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	userId := claims.Sub

	data, err := t.service.GetUserData(r.Context(), userId)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	userData = UserDataResponse{
		Id:         data.Id,
		UUID:       data.UUID,
		Login:      data.Login,
		Email:      data.Email,
		TelegramId: data.TelegramId,
		CreatedAt:  data.CreatedAt.Format("2006-01-02"),
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(userData)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerTelegramBind(w http.ResponseWriter, r *http.Request) {
	var request TelegramBindRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &request) {
		return
	}

	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	if err = t.service.BindTelegram(r.Context(), &request.TelegramId, claims.Sub); err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerTelegramDelete(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	if err := t.service.DeleteTelegram(r.Context(), claims.Sub); err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (t *Transport) handlerAuthHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}

	userId := claims.Sub
	authHistory, err := t.service.GetAuthHistory(r.Context(), userId)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
	
	var response UserAuthHistoryResponse
	if authHistory != nil {
		for _, entry := range authHistory {
			userHist := UserAuthHistoryResponseItem{
				Id:        entry.Id,
				Agent:     entry.Agent,
				Ip:        entry.Ip,
				Timestamp: entry.Timestamp.Format("2006.01.02 15:04:05"),
			}
			response.Items = append(response.Items, userHist)
		}
	} else {
		response.Items = nil
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}
