package scheduler

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/Uikola/earn-it/internal/timeutil"
)

func (s *Scheduler) processWeeklyBonuses() {
	ctx := context.Background()

	users, err := s.userRepository.Users(ctx)
	if err != nil {
		slog.Error("failed to fetch users", slog.String("error", err.Error()))
		return
	}

	now := time.Now()

	for _, user := range users {
		s.processUserWeeklyBonus(ctx, user.ID, user.Timezone, user.RewardWeeklyBonus, now)
	}
}

func (s *Scheduler) processUserWeeklyBonus(ctx context.Context, userID int64, timezone string, rewardWeeklyBonus int32, now time.Time) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		slog.Error("invalid timezone", slog.Int64("user_id", userID), slog.String("timezone", timezone), slog.String("error", err.Error()))
		return
	}

	weekStart := timeutil.WeekStart(now, loc)

	_, err = s.habitRepository.WeeklyBonus(ctx, userID, weekStart)
	if err == nil {
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to check weekly bonus", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return
	}

	habits, err := s.habitRepository.HabitsByUserID(ctx, userID)
	if err != nil {
		slog.Error("failed to fetch habits", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return
	}

	if len(habits) == 0 {
		return
	}

	for _, habit := range habits {
		logs, err := s.habitRepository.HabitLogsForWeek(ctx, habit.ID, weekStart)
		if err != nil {
			slog.Error("failed to fetch habit logs", slog.Int64("user_id", userID), slog.Int64("habit_id", habit.ID), slog.String("error", err.Error()))
			return
		}

		if int32(len(logs)) < habit.WeaklyGoal {
			return
		}
	}

	err = s.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		if err := s.habitRepository.CreateWeeklyBonus(txctx, userID, weekStart); err != nil {
			return err
		}

		user, err := s.userRepository.UserByID(txctx, userID)
		if err != nil {
			return err
		}

		user.Balance += rewardWeeklyBonus
		return s.userRepository.UpdateUser(txctx, user)
	})
	if err != nil {
		slog.Error("failed to award weekly bonus", slog.Int64("user_id", userID), slog.String("error", err.Error()))
		return
	}

	slog.Info("weekly bonus awarded", slog.Int64("user_id", userID), slog.Int("amount", int(rewardWeeklyBonus)))

	if err := s.notifier.Notify(userID, "weekly_bonus_notification", struct{ Amount int32 }{rewardWeeklyBonus}); err != nil {
		slog.Error("failed to send notification", slog.Int64("user_id", userID), slog.String("error", err.Error()))
	}
}
