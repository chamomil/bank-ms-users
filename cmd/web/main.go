package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
	"x-bank-users/config"
	"x-bank-users/core/web"
	"x-bank-users/infra/gmail"
	"x-bank-users/infra/hasher"
	"x-bank-users/infra/postgres"
	"x-bank-users/infra/random"
	"x-bank-users/infra/redis"
	"x-bank-users/infra/telegram"
	"x-bank-users/transport/http"
	"x-bank-users/transport/http/jwt"
)

var (
	addr       = flag.String("addr", ":8080", "")
	configFile = flag.String("config", "config.json", "")
)

func main() {
	f, perr := os.Create("cpu.pprof")
	if perr != nil {
		log.Fatal(perr)
		return
	}
	err := pprof.StartCPUProfile(f)
	if err != nil {
		return
	}
	defer pprof.StopCPUProfile()
	flag.Parse()
	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	passwordHasher := hasher.NewService()

	jwtHs512, err := jwt.NewHS512(conf.Hs512SecretKey)
	if err != nil {
		log.Fatal(err)
	}
	//jwtRs256, err := jwt.NewRS256(conf.Rs256PrivateKey, conf.Rs256PublicKey)
	//if err != nil {
	//	log.Fatal(err)
	//}

	redisService, err := redis.NewService(conf.Redis.Password, conf.Redis.Host, conf.Redis.Port, conf.Redis.Database, conf.Redis.MaxCons)
	if err != nil {
		log.Fatal(err)
	}
	gmailService := gmail.NewService(conf.Gmail.Host, conf.Gmail.SenderName, conf.Gmail.SenderEmail, conf.Gmail.Login, conf.Gmail.Password, conf.Gmail.UrlToActivate, conf.Gmail.UrlToRestore)

	randomGenerator := random.NewService()

	postgresService, err := postgres.NewService(conf.Postgres.Login, conf.Postgres.Password, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DataBase, conf.Postgres.MaxCons)
	if err != nil {
		log.Fatal(err)
	}

	telegramService := telegram.NewService(conf.Telegram.BaseURL, conf.Telegram.Login, conf.Telegram.Password)
	service := web.NewService(&postgresService, &randomGenerator, &redisService, &gmailService, &passwordHasher, &redisService, &redisService, &telegramService, &redisService)

	transport := http.NewTransport(service, &jwtHs512)

	errCh := transport.Start(*addr)
	interruptsCh := make(chan os.Signal, 1)
	signal.Notify(interruptsCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-errCh:
		log.Fatal(err)
	case <-interruptsCh:
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		err = transport.Stop(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
