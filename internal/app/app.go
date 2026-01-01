package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/koan6gi/notifier/internal/config"
	"github.com/koan6gi/notifier/internal/repository"
	"github.com/koan6gi/notifier/internal/service"
	"github.com/koan6gi/notifier/internal/transport/rest"
	v1 "github.com/koan6gi/notifier/internal/transport/rest/v1"
	"github.com/koan6gi/notifier/internal/updater"
	"github.com/koan6gi/notifier/pkg/logger"

	"go.uber.org/zap"
)

const (
	configPath = "config/config.yaml"

	timeout = 30 * time.Second
)

func Run() error {
	cfg, err := config.Parse(configPath)
	if err != nil {
		return fmt.Errorf("config error: %v", err)
	}

	lg, err := logger.New()
	if err != nil {
		return err
	}
	defer lg.Sync()

	repo := repository.New()

	svc := service.New(repo)

	srv := rest.NewServer(&cfg.Server)

	hh := v1.NewHandlers(svc)

	srv.RegisterHandlers(hh, lg)

	worker := updater.New(&cfg.Request, repo)

	loggerCtx := logger.WithContext(context.Background(), lg)

	ctx, cancel := context.WithCancel(loggerCtx)
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		lg.Info("worker started")
		if err := worker.Run(ctx); err != nil {
			lg.Error("worker error", zap.Error(err))
		}

		lg.Info("worker stopped")
	})

	wg.Go(func() {
		lg.Info("server started", zap.Int("port", cfg.Server.Port))
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			lg.Error("server error", zap.Error(err))
		}

		lg.Info("server stopped")
	})

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	sig := <-signalCh
	lg.Info("recieved signal, shutting down", zap.String("signal", sig.String()))

	graceSh, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	srv.Shutdown(graceSh)

	cancel()

	wg.Wait()

	return nil
}
