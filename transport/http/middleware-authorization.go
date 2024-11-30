package http

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func (t *Transport) authMiddleware(allow2Fa bool) middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				t.errorHandler.setUnauthorizedError(w, errors.New("отсутствует заголовок Authorization"))
				return
			}
			parts := strings.Split(header, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				t.errorHandler.setUnauthorizedError(w, errors.New("неверный формат заголовка Authorization"))
				return
			}
			token := parts[1]
			claims, err := t.authorizer.VerifyAuthorization(r.Context(), []byte(token))
			if err != nil {
				t.errorHandler.setUnauthorizedError(w, err)
				return
			}

			if !allow2Fa && claims.Is2FAToken {
				t.errorHandler.setUnauthorizedError(w, errors.New("требуется 2FA"))
				return
			}

			ctx := context.WithValue(r.Context(), t.claimsCtxKey, &claims)
			handlerFunc(w, r.WithContext(ctx))
		}
	}
}
