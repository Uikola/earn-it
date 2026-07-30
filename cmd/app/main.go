package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/Uikola/keepflame/internal/handlers"
	"github.com/Uikola/keepflame/internal/handlers/auth"
	"github.com/Uikola/keepflame/internal/repository/postgres"
	"github.com/Uikola/keepflame/internal/repository/postgres/refreshtoken"
	"github.com/Uikola/keepflame/internal/repository/postgres/user"
	authservice "github.com/Uikola/keepflame/internal/service/auth"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	db, err := postgres.InitDB(ctx, "postgres://postgres:password@localhost:5433/keepflame")
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer db.Close()

	userRepo := user.NewRepository(db)
	refreshTokenRepo := refreshtoken.NewRepository(db)

	jwtSecret := []byte("your-secret-key-change-in-production")
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	authService := authservice.NewService(userRepo, refreshTokenRepo, jwtSecret, accessTTL, refreshTTL)

	authHandler := auth.NewHandler(authService, false, int(refreshTTL.Seconds()))

	rootHandler := &handlers.Root{
		AuthHandler: *authHandler,
	}

	handler, err := handlers.NewServer(rootHandler)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	httpServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Addr:         ":8080",
		Handler:      handler,
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("starting HTTP server", slog.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server failed: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			slog.Info("received signal starting shutdown", slog.String("signal", sig.String()))
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				slog.Info("HTTP server shutdown error", slog.String("error", err.Error()))
				return fmt.Errorf("shutdown HTTP server: %w", err)
			}
			slog.Info("HTTP server gracefully stopped")
			return nil
		case <-ctx.Done():
			slog.Info("context cancelled, exiting signal handler")
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
