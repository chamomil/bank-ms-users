package http

import (
	"net/http"
	_ "net/http/pprof"
)

func (t *Transport) routes() http.Handler {
	corsHandler := t.corsHandler("*", "*", "*", "")
	corsMiddleware := t.corsMiddleware(corsHandler)

	defaultMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
	}

	signIn2FaMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
		t.authMiddleware(true),
	}

	userMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
		t.authMiddleware(false),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultMiddlewareGroup.Apply(t.handlerNotFound))
	mux.HandleFunc("OPTIONS /", corsHandler)

	mux.HandleFunc("POST /v1/auth/sign-up", defaultMiddlewareGroup.Apply(t.handlerSignUp))
	mux.HandleFunc("POST /v1/auth/sign-in", defaultMiddlewareGroup.Apply(t.handlerSignIn))
	mux.HandleFunc("POST /v1/auth/sign-in/2fa", signIn2FaMiddlewareGroup.Apply(t.handlerSignIn2FA))
	mux.HandleFunc("POST /v1/auth/refresh", defaultMiddlewareGroup.Apply(t.handlerRefresh))

	mux.HandleFunc("GET /v1/countries", userMiddlewareGroup.Apply(t.handlerGetCountries))

	mux.HandleFunc("GET /v1/me/personal-data", userMiddlewareGroup.Apply(t.handlerGetUserPersonalData))
	mux.HandleFunc("PUT /v1/me/personal-data", userMiddlewareGroup.Apply(t.handlerAddUserPersonalData))
	mux.HandleFunc("GET /v1/me", userMiddlewareGroup.Apply(t.handlerGetUserData))
	mux.HandleFunc("GET /v1/me/auth-history", userMiddlewareGroup.Apply(t.handlerAuthHistory))
	mux.HandleFunc("GET /v1/me/work", userMiddlewareGroup.Apply(t.handlerGetWorkplaces))
	mux.HandleFunc("POST /v1/me/work", userMiddlewareGroup.Apply(t.handlerAddWorkplace))

	return mux
}
