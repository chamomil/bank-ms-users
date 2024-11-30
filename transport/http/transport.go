package http

import (
	"context"
	"net/http"
	"x-bank-users/auth"
	"x-bank-users/cerrors"
	"x-bank-users/core/web"
	"x-bank-users/ercodes"
)

type (
	Transport struct {
		service      web.Service
		authorizer   auth.Authorizer
		errorHandler errorHandler

		srv *http.Server

		claimsCtxKey string
	}
)

func NewTransport(service web.Service, authorizer auth.Authorizer) Transport {
	return Transport{
		service:    service,
		authorizer: authorizer,
		errorHandler: errorHandler{
			defaultStatusCode: http.StatusBadRequest,
			statusCodes: map[cerrors.Code]int{
				ercodes.BcryptHashing: http.StatusInternalServerError,
			},
		},
		claimsCtxKey: "CLAIMS",
	}
}

func (t *Transport) Start(addr string) chan error {
	t.srv = &http.Server{Addr: addr, Handler: t.routes()}
	ch := make(chan error)

	go func() {
		ch <- t.srv.ListenAndServe()
	}()

	return ch
}

func (t *Transport) Stop(ctx context.Context) error {
	return t.srv.Shutdown(ctx)
}
