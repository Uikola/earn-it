package handlers

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/sqlc"
	"github.com/google/martian/log"
	"github.com/nlypage/intele"
	"github.com/nlypage/intele/collector"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type habitRepository interface {
	HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error)
	HabitByID(ctx context.Context, habitID int64) (models.Habit, error)
	CreateHabit(ctx context.Context, userID int64, name string, weaklyGoal int32, rewardPerExecute int32) (models.Habit, error)
	DeleteHabit(ctx context.Context, habitID int64) error

	CreateHabitLog(ctx context.Context, habitID int64) error
	HabitLogsForWeek(ctx context.Context, habitID int64, timezoneStr string) ([]sqlc.HabitLog, error)
}

type hUserRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
}

type HabitsHandler struct {
	layout *layout.Layout
	input  *intele.InputManager

	transactor       postgres.Transactor
	habitsRepository habitRepository
	userRepository   hUserRepository
}

func NewHabitsHandler(
	layout *layout.Layout,
	input *intele.InputManager,
	transactor postgres.Transactor,
	habitsRepository habitRepository,
	userRepository hUserRepository,
) *HabitsHandler {
	return &HabitsHandler{
		layout: layout,
		input:  input,

		transactor:       transactor,
		habitsRepository: habitsRepository,
		userRepository:   userRepository,
	}
}

type habitToPrint struct {
	Name                string
	WeaklyGoalDone      int32
	WeaklyGoal          int32
	WeaklyGoalCompleted bool
	RewardPerExecute    int32
}

func (h *HabitsHandler) Habits(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	return c.Edit(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, user.Timezone, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}

func (h *HabitsHandler) NewHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "send")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	inputCollector := collector.New()
	inputCollector.Collect(c.Message())

	isFirst := true

	var steps []struct {
		promptKey string
		result    *string
	}

	steps = []struct {
		promptKey string
		result    *string
	}{
		{
			promptKey: "input_habit_name",
			result:    new(string),
		},

		{
			promptKey: "input_habit_weakly_goal",
			result:    new(string),
		},

		{
			promptKey: "input_habit_reward_per_execute",
			result:    new(string),
		},
	}

	for _, step := range steps {
		done := false

		markup := h.layout.Markup(c, "habitsMenuBack")

		if isFirst {
			_ = c.Edit(
				h.layout.Text(c, step.promptKey),
				markup,
			)
		} else {
			_ = inputCollector.Send(c,
				h.layout.Text(c, step.promptKey),
				markup,
			)
		}
		isFirst = false

		for !done {
			response, errGet := h.input.Get(context.Background(), c.Sender().ID, 0, nil)
			if response.Message != nil {
				inputCollector.Collect(response.Message)
			}
			switch {
			case response.Canceled:
				_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
				return nil
			case errGet != nil:
				log.Errorf("(user: %d) error while input step (%s): %v", c.Sender().ID, step.promptKey, errGet)
				_ = inputCollector.Send(c,
					h.layout.Text(c, "technical_issues"),
					markup,
				)
			case response.Callback != nil:
				*step.result = ""
				_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
				done = true
			default:
				*step.result = response.Message.Text
				_ = inputCollector.Clear(c, collector.ClearOptions{IgnoreErrors: true})
				done = true
			}
		}
	}

	var (
		habitName             string
		habitWeaklyGoal       int32
		habitRewardForExecute int32
	)

	habitName = *steps[0].result

	// TODO: Сделать на это всё валидацию
	habitWeaklyGoalTemp, _ := strconv.ParseInt(*steps[1].result, 10, 32)
	habitWeaklyGoal = int32(habitWeaklyGoalTemp)

	// TODO: Сделать на это всё валидацию
	habitRewardForExecuteTemp, _ := strconv.ParseInt(*steps[2].result, 10, 32)
	habitRewardForExecute = int32(habitRewardForExecuteTemp)

	_, err := h.habitsRepository.CreateHabit(ctx, userID, habitName, habitWeaklyGoal, habitRewardForExecute)
	if err != nil {
		log.Errorf("failed to create habit: %v", err)
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Send(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, user.Timezone, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}

func (h *HabitsHandler) CompleteHabits(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	var rows []tele.Row

	markup := c.Bot().NewMarkup()
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, user.Timezone)
		if err != nil {
			log.Errorf("failed to fetch habit logs for week: %v", err)
			return c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			)
		}

		rows = append(rows, markup.Row(*h.layout.Button(c, "habits:complete:habit", struct {
			ID                  int64
			Name                string
			WeaklyGoalDone      int32
			WeaklyGoal          int32
			WeaklyGoalCompleted bool
		}{
			ID:                  habit.ID,
			Name:                habit.Name,
			WeaklyGoalDone:      int32(len(habitLogsForWeek)),
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeaklyGoal,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "habitsMenuBack")),
	)
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "habit_complete_text"),
		markup,
	)
}

func (h *HabitsHandler) CompleteHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	habitID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		log.Errorf("invalid callback data for complete habit handler: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	habit, err := h.habitsRepository.HabitByID(ctx, habitID)
	if err != nil {
		log.Errorf("failed to fetch habit: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	err = h.transactor.WithinTransaction(ctx, func(txctx context.Context) error {
		if err := h.habitsRepository.CreateHabitLog(txctx, habitID); err != nil {
			return fmt.Errorf("failed to create habit: %v", err)
		}

		user.Balance += habit.RewardPerExecute
		if err := h.userRepository.UpdateUser(ctx, *user); err != nil {
			return fmt.Errorf("failed to update user: %v", err)
		}
		return nil
	})
	if err != nil {
		log.Errorf("failed to complete transaction: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Edit(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, user.Timezone, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}

func (h *HabitsHandler) DeleteHabits(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	var rows []tele.Row

	markup := c.Bot().NewMarkup()
	for _, habit := range habits {
		rows = append(rows, markup.Row(*h.layout.Button(c, "habits:delete:habit", struct {
			ID   int64
			Name string
		}{
			ID:   habit.ID,
			Name: habit.Name,
		})))
	}

	rows = append(
		rows,
		markup.Row(*h.layout.Button(c, "habitsMenuBack")),
	)
	markup.Inline(rows...)

	return c.Edit(
		h.layout.Text(c, "habit_delete_text"),
		markup,
	)
}

func (h *HabitsHandler) DeleteHabit(c tele.Context) error {
	ctx := context.Background()
	userID := c.Sender().ID

	user := h.userByIDWithProcessedError(ctx, c, userID, "edit")
	if user == nil {
		return nil
	}

	habits := h.habitsByUserIDWithProcessedError(ctx, c, userID, "edit")
	if habits == nil {
		return nil
	}

	habitID, err := strconv.ParseInt(c.Callback().Data, 10, 64)
	if err != nil {
		log.Errorf("invalid callback data for complete habit handler: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	if err := h.habitsRepository.DeleteHabit(ctx, habitID); err != nil {
		log.Errorf("failed to delete habit log: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	return c.Edit(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            h.habitsToPrint(ctx, c, user.Timezone, habits),
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}

func (h *HabitsHandler) userByIDWithProcessedError(ctx context.Context, c tele.Context, userID int64, action string) *models.User {
	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch user: %v", err)

		switch action {
		case "send":
			if err := c.Send(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("SendError while userByIDWithProcessedError")
				return nil
			}
		case "edit":
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while userByIDWithProcessedError")
				return nil
			}
		}

		return nil
	}

	return &user
}

func (h *HabitsHandler) habitsByUserIDWithProcessedError(ctx context.Context, c tele.Context, userID int64, action string) []models.Habit {
	habits, err := h.habitsRepository.HabitsByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch habits: %v", err)

		switch action {
		case "send":
			if err := c.Send(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("SendError while habitsByUserIDWithProcessedError")
				return nil
			}
		case "edit":
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while habitsByUserIDWithProcessedError")
				return nil
			}
		}
	}

	return habits
}

func (h *HabitsHandler) habitsToPrint(ctx context.Context, c tele.Context, timezone string, habits []models.Habit) []habitToPrint {
	habitsToPrint := make([]habitToPrint, 0, len(habits))
	for _, habit := range habits {
		habitLogsForWeek, err := h.habitsRepository.HabitLogsForWeek(ctx, habit.ID, timezone)
		if err != nil {
			log.Errorf("failed to fetch habit logs for week: %v", err)
			if err := c.Edit(
				h.layout.Text(c, "technical_issues"),
				h.layout.Markup(c, "mainMenuBack"),
			); err != nil {
				log.Errorf("EditError while habitsToPrint")
				return nil
			}
		}

		habitsToPrint = append(habitsToPrint, habitToPrint{
			Name:                habit.Name,
			WeaklyGoalDone:      int32(len(habitLogsForWeek)),
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: int32(len(habitLogsForWeek)) >= habit.WeaklyGoal,
			RewardPerExecute:    habit.RewardPerExecute,
		})
	}

	return habitsToPrint
}
