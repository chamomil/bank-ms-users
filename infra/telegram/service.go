package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"x-bank-users/cerrors"
	"x-bank-users/ercodes"
)

type (
	Service struct {
		client   *http.Client
		baseURL  string
		login    string
		password string
	}
)

func NewService(baseURL, Login, Password string) Service {
	return Service{
		client:   &http.Client{},
		baseURL:  baseURL,
		login:    Login,
		password: Password,
	}
}

func (s *Service) Send2FaCode(ctx context.Context, telegramId int64, code string) error {
	reqBody, err := json.Marshal(map[string]interface{}{
		"userId": telegramId,
		"code":   code,
	})
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.TelegramSendError, err, "Ошибка отправки кода")
	}

	url := s.baseURL + "/internal/v1/2fa"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.TelegramSendError, err, "Ошибка отправки кода")
	}

	req.SetBasicAuth(s.login, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.TelegramSendError, err, "Ошибка отправки кода")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return cerrors.NewErrorWithUserMessage(ercodes.TelegramSendError, nil, "Ошибка отправки кода")
	}

	return nil
}
