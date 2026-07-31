package handlers

import (
	"context"
	"strconv"

	"github.com/Uikola/earn-it/internal/models"
	"github.com/google/martian/log"
	"github.com/nlypage/intele"
	"github.com/nlypage/intele/collector"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

type habitRepository interface {
	HabitsByUserID(ctx context.Context, userID int64) ([]models.Habit, error)
	CreateHabit(ctx context.Context, userID int64, name string, weaklyGoal int32, rewardPerExecute int32) (models.Habit, error)
}

type hUserRepository interface {
	UserByID(ctx context.Context, id int64) (models.User, error)
}

type HabitsHandler struct {
	layout *layout.Layout
	input  *intele.InputManager

	habitsRepository habitRepository
	userRepository   hUserRepository
}

func NewHabitsHandler(layout *layout.Layout, input *intele.InputManager, habitsRepository habitRepository, userRepository hUserRepository) *HabitsHandler {
	return &HabitsHandler{
		layout:           layout,
		input:            input,
		habitsRepository: habitsRepository,
		userRepository:   userRepository,
	}
}

func (h *HabitsHandler) Habits(c tele.Context) error {
	ctx := context.Background()

	userID := c.Sender().ID

	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch user: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	habits, err := h.habitsRepository.HabitsByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch habits: %v", err)
		return c.Edit(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	type habitToPrint struct {
		Name                string
		WeaklyGoalDone      int32
		WeaklyGoal          int32
		WeaklyGoalCompleted bool
		RewardPerExecute    int32
	}

	habitsToPrint := make([]habitToPrint, 0, len(habits))
	for _, habit := range habits {
		habitsToPrint = append(habitsToPrint, habitToPrint{
			Name:                habit.Name,
			WeaklyGoalDone:      0, // TODO: Пока сложно, надо написать метод подсчета HabitLog'а для repo
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: false, // TODO: Когда появиться WeaklyGoalDone, то добавить сюда WeaklyGoalDone == WeaklyGoal
			RewardPerExecute:    habit.RewardPerExecute,
		})
	}

	return c.Edit(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            habitsToPrint,
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}

func (h *HabitsHandler) NewHabit(c tele.Context) error {
	ctx := context.Background()

	userID := c.Sender().ID

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

	// TODO: Убрать дубли
	user, err := h.userRepository.UserByID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch user: %v", err)
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	habits, err := h.habitsRepository.HabitsByUserID(ctx, userID)
	if err != nil {
		log.Errorf("failed to fetch habits: %v", err)
		return c.Send(
			h.layout.Text(c, "technical_issues"),
			h.layout.Markup(c, "mainMenuBack"),
		)
	}

	type habitToPrint struct {
		Name                string
		WeaklyGoalDone      int32
		WeaklyGoal          int32
		WeaklyGoalCompleted bool
		RewardPerExecute    int32
	}

	habitsToPrint := make([]habitToPrint, 0, len(habits))
	for _, habit := range habits {
		habitsToPrint = append(habitsToPrint, habitToPrint{
			Name:                habit.Name,
			WeaklyGoalDone:      0, // TODO: Пока сложно, надо написать метод подсчета HabitLog'а для repo
			WeaklyGoal:          habit.WeaklyGoal,
			WeaklyGoalCompleted: false, // TODO: Когда появиться WeaklyGoalDone, то добавить сюда WeaklyGoalDone == WeaklyGoal
			RewardPerExecute:    habit.RewardPerExecute,
		})
	}

	return c.Send(
		h.layout.Text(c, "habits_menu_text", struct {
			Habits            []habitToPrint
			RewardWeeklyBonus int32
		}{
			Habits:            habitsToPrint,
			RewardWeeklyBonus: user.RewardWeeklyBonus,
		}),
		h.layout.Markup(c, "habitsMenu"),
	)
}
