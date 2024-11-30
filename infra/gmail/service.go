package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"x-bank-users/cerrors"
	"x-bank-users/ercodes"
)

type (
	Service struct {
		host          string
		senderName    string
		senderEmail   string
		auth          loginAuth
		urlToActivate string
		urlToRestore  string
	}
)

func NewService(host, senderName, senderEmail, login, password, urlToActivate, urlToRestore string) Service {
	return Service{
		host:        host,
		senderName:  senderName,
		senderEmail: senderEmail,
		auth: loginAuth{
			username: login,
			password: password,
		},
		urlToActivate: urlToActivate,
		urlToRestore:  urlToRestore,
	}
}

func (s *Service) send(email, subject, message string) error {
	header := fmt.Sprintf(
		"To: %s\r\n"+
			"From: %s <%s>\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\nContent-Type: text/html; charset=utf-8\r\n\r\n",
		email, s.mailUTF8(s.senderName), s.senderEmail, s.mailUTF8(subject))

	return smtp.SendMail(s.host, &s.auth, s.senderEmail, []string{email}, []byte(header+message))
}

func (s *Service) mailUTF8(data string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(data)) + "?="
}

func (s *Service) SendActivationCode(_ context.Context, email, code string) error {
	err := s.send(email, "Активация аккаунта X-BANK", fmt.Sprintf("Для активации перейдите по ссылке - %s?code=%s", s.urlToActivate, code))
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.GmailSendError, err, "Ошибка отправки письма")
	}

	return nil
}

func (s *Service) SendRecoveryCode(_ context.Context, email, code string) error {
	err := s.send(email, "Восстановление аккаунта X-BANK", fmt.Sprintf("Для активации перейдите по ссылке - %s?code=%s", s.urlToActivate, code))
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.GmailSendError, err, "Ошибка отправки письма")
	}

	return nil
}
