package telegram

import (
	"strings"

	"github.com/Uikola/earn-it/internal/repository/postgres"
	"github.com/Uikola/earn-it/internal/repository/postgres/habit"
	"github.com/Uikola/earn-it/internal/repository/postgres/user"
	"github.com/Uikola/earn-it/internal/repository/redis"
	"github.com/Uikola/earn-it/internal/telegram/handlers"
	"github.com/Uikola/earn-it/internal/telegram/handlers/habits"
	"github.com/nlypage/intele"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
	"gopkg.in/telebot.v3/middleware"
)

type Bot struct {
	*tele.Bot
	Layout *layout.Layout
	Input  *intele.InputManager
}

func NewBot(redisClient *redis.StateRepository) (*Bot, error) {
	lt, err := layout.New("telegram.yml")
	if err != nil {
		return nil, err
	}

	settings := lt.Settings()

	b, err := tele.NewBot(settings)
	if err != nil {
		return nil, err
	}

	if cmds := lt.Commands(); cmds != nil {
		if err = b.SetCommands(cmds); err != nil {
			return nil, err
		}
	}

	bot := &Bot{
		Bot:    b,
		Layout: lt,
		Input: intele.NewInputManager(intele.InputOptions{
			Storage: redisClient,
		}),
	}

	return bot, nil
}

func (bot *Bot) Setup(
	transactor postgres.Transactor,
	userRepository *user.Repository,
	habitRepository *habit.Repository,
) {
	startHandler := handlers.NewStartHandler(bot.Layout, userRepository)
	habitsHandler := habits.NewHandler(bot.Layout, bot.Input, transactor, habitRepository, userRepository)

	bot.Use(bot.Layout.Middleware("ru"))
	bot.Use(middleware.AutoRespond())
	bot.Handle(tele.OnText, bot.Input.MessageHandler())
	bot.Handle(tele.OnMedia, bot.Input.MessageHandler())
	bot.Handle(tele.OnCallback, bot.Input.CallbackHandler())
	bot.Handle(tele.OnVideoNote, bot.Input.MessageHandler())
	bot.Use(bot.ResetInputOnBack)
	bot.Handle(bot.Layout.Callback("core:hide"), hide)
	bot.Handle(bot.Layout.Callback("core:cancel"), hide)
	bot.Handle(bot.Layout.Callback("core:back"), hide)

	// Setup handlers
	// Start
	bot.Handle("/start", startHandler.Start)
	bot.Handle(bot.Layout.Callback("mainMenuBack"), startHandler.MainMenu)

	bot.Handle(bot.Layout.Callback("habitsMenu"), habitsHandler.Habits)
	bot.Handle(bot.Layout.Callback("habitsMenuBack"), habitsHandler.Habits)
	bot.Handle(bot.Layout.Callback("habits:new"), habitsHandler.NewHabit)
	bot.Handle(bot.Layout.Callback("habits:complete"), habitsHandler.CompleteHabits)
	bot.Handle(bot.Layout.Callback("habits:complete:habit"), habitsHandler.CompleteHabit)
	bot.Handle(bot.Layout.Callback("habits:delete"), habitsHandler.DeleteHabits)
	bot.Handle(bot.Layout.Callback("habits:delete:habit"), habitsHandler.DeleteHabit)
}

// ResetInputOnBack middleware clears the input state when the back button is pressed.
func (bot *Bot) ResetInputOnBack(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if c.Callback() != nil {
			if strings.Contains(c.Callback().Data, "back") || strings.Contains(c.Callback().Unique, "back") {
				bot.Input.Cancel(c.Sender().ID)
			}
		}
		if c.Message() != nil {
			if strings.HasPrefix(c.Message().Text, "/") {
				bot.Input.Cancel(c.Sender().ID)
			}
		}

		return next(c)
	}
}

func hide(c tele.Context) error {
	return c.Delete()
}
