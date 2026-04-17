package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	apiv3 "github.com/bmstu-itstech/itsreg/internal/api/v3"
	"github.com/bmstu-itstech/itsreg/internal/api/v3/jwtauth"
	"github.com/bmstu-itstech/itsreg/internal/app"
	"github.com/bmstu-itstech/itsreg/internal/app/bootstrap"
	"github.com/bmstu-itstech/itsreg/internal/app/dispatcher"
	"github.com/bmstu-itstech/itsreg/internal/app/port"
	"github.com/bmstu-itstech/itsreg/internal/config"
	"github.com/bmstu-itstech/itsreg/internal/infra/inmemory"
	"github.com/bmstu-itstech/itsreg/internal/infra/jwt"
	"github.com/bmstu-itstech/itsreg/internal/infra/postgres"
	"github.com/bmstu-itstech/itsreg/internal/infra/telegram"
	"github.com/bmstu-itstech/itsreg/pkg/logs"
	"github.com/bmstu-itstech/itsreg/pkg/logs/sl"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

const apiPrefix = "/api/v3"

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()
	if cfgPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	cfg := config.MustLoad(cfgPath)

	l := logs.NewLogger(cfg.Logging.Level)

	l.Debug(fmt.Sprintf("config: %+v", cfg))

	repos := postgres.MustNewRepository(cfg.Postgres)
	bus := inmemory.NewEventBus(l)
	inbound := dispatcher.NewInboundDispatcher(l)
	instanceManager := telegram.NewInstanceManager(inbound, l)
	sender := telegram.NewMessageSender(l)
	tokenService := jwt.MustNewTokenService(cfg.JWT)

	infra := app.Infra{
		BotMetaProvider:      repos,
		BotRepository:        repos,
		EventBus:             bus,
		InstanceManager:      instanceManager,
		MessageSender:        sender,
		RunRepository:        repos,
		ScriptMetaProvider:   repos,
		ScriptRepository:     repos,
		ThreadRepository:     repos,
		ThreadsTableProvider: repos,
		UserRepository:       repos,
	}
	a := app.NewApplication(infra, l)

	// Ленивое связывание

	inbound.SetEntryHandler(a.Commands.Entry)
	inbound.SetProcessHandler(a.Commands.Process)

	// Регистрация обработчиков событий

	mustSubscribe(bus, "run.start_requested", a.Events.StartOnRunStartRequested)
	mustSubscribe(bus, "run.recover_requested", a.Events.StartOnRunRecoverRequested)

	// Восстановление состояние

	recoverer := bootstrap.NewRecoverActiveRunsHandler(repos, bus, l)
	if err := recoverer.Recover(context.Background()); err != nil {
		l.Error("error recovering active runs", "error", err)
		os.Exit(1)
	}

	// Инициализация HTTP сервера

	root := chi.NewRouter()
	root.Use(middleware.RealIP)
	root.Use(sl.NewLoggerMiddleware(l))
	root.Use(middleware.Recoverer)
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.HTTP.CORSAllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           cfg.HTTP.CORSMaxAge,
	})
	root.Use(corsMiddleware.Handler)
	root.Use(
		middleware.SetHeader("X-Content-Type-Options", "nosniff"),
		middleware.SetHeader("X-Frame-Options", "deny"),
	)
	root.Use(middleware.NoCache)
	root.Use(jwtauth.NewMiddleware(tokenService).Handler)
	s := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler: apiv3.HandlerFromMuxWithBaseURL(apiv3.NewServer(a, apiPrefix), root, apiPrefix),
	}

	// Асинхронный запуск нескольких серверов.
	// Так, добавление нового сервера (например, читающего из брокера), осуществляется запуском ещё одной горутины
	// с записью ошибки в errCh.

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	errCh := make(chan error, 1)

	go func() {
		l.Info("starting http server", slog.String("addr", s.Addr))
		err := s.ListenAndServe()
		errCh <- err
	}()

	var err error
	select {
	case <-ctx.Done():
		l.Info("received cancel signal, gracefully shutting down")
		err = s.Shutdown(context.Background())
		if err != nil {
			l.Error("error shutting down http server", "error", err)
		}
	case err = <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			l.Error("listen error", slog.String("error", err.Error()))
			cancel()
		}
	}
}

func mustSubscribe(bus port.EventBus, eventName string, h port.EventHandler) {
	if err := bus.Subscribe(eventName, h); err != nil {
		panic(err)
	}
}
