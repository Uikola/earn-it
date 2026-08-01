package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/habit"
	"github.com/Uikola/earn-it/internal/repository/postgres/user"
	"github.com/Uikola/earn-it/internal/repository/redis"
	"github.com/Uikola/earn-it/internal/telegram"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	db, err := postgres.InitDB(ctx, "postgres://postgres:password@localhost:5433/earnit?sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer db.Close()

	redisClient, err := redis.InitClient(ctx, "localhost:6379", "password")
	if err != nil {
		return fmt.Errorf("failed to init redis client: %w", err)
	}
	defer redisClient.Close()

	stateRepository := redis.NewStateRepository(redisClient)

	transactor := postgres.NewPgxTransactor(db)
	userRepository := user.NewRepository(db)
	habitRepository := habit.NewRepository(db)

	bot, err := telegram.NewBot(stateRepository)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	bot.Setup(transactor, userRepository, habitRepository)

	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("bot panic: %v", r)
			}
		}()
		bot.Start()
	}()

	log.Println("Bot started")

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-sigCh:
			slog.Info("received signal", slog.String("signal", sig.String()))
			return
		case botErr := <-errChan:
			slog.Info("bot error", slog.String("error", botErr.Error()))
			return
		case <-ctx.Done():
			slog.Info("context cancelled, exiting signal handler")
			return
		}
	})

	wg.Wait()

	return nil
}
