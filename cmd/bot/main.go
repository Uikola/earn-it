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

	"github.com/joho/godotenv"

	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/habit"
	"github.com/Uikola/earn-it/internal/repository/postgres/shop"
	"github.com/Uikola/earn-it/internal/repository/postgres/task"
	"github.com/Uikola/earn-it/internal/repository/postgres/transaction"
	"github.com/Uikola/earn-it/internal/repository/postgres/user"
	"github.com/Uikola/earn-it/internal/repository/redis"
	"github.com/Uikola/earn-it/internal/scheduler"
	"github.com/Uikola/earn-it/internal/telegram"
	"github.com/Uikola/earn-it/internal/telegram/notifier"
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

	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := postgres.InitDB(ctx, postgresDSN)
	if err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer db.Close()

	redisAddr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient, err := redis.InitClient(ctx, redisAddr, redisPassword)
	if err != nil {
		return fmt.Errorf("failed to init redis client: %w", err)
	}
	defer redisClient.Close()

	stateRepository := redis.NewStateRepository(redisClient)

	transactor := postgres.NewPgxTransactor(db)
	userRepository := user.NewRepository(db)
	habitRepository := habit.NewRepository(db)
	taskRepository := task.NewRepository(db)
	transactionRepository := transaction.NewRepository(db)
	shopRepository := shop.NewRepository(db)

	bot, err := telegram.NewBot(stateRepository)
	if err != nil {
		return fmt.Errorf("failed to create bot: %w", err)
	}

	notif := notifier.New(bot.Bot, bot.Layout)

	sched := scheduler.New(transactor, userRepository, habitRepository, transactionRepository, notif)
	sched.Start()
	defer sched.Stop()

	bot.Setup(transactor, userRepository, habitRepository, taskRepository, transactionRepository, shopRepository)

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
