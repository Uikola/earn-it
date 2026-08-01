package helpers

import (
	"context"
	"errors"

	"github.com/google/martian/log"
	"github.com/nlypage/intele"
	"github.com/nlypage/intele/collector"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/layout"
)

// InputStep описывает один шаг ввода
type InputStep struct {
	Name      string // идентификатор шага (для сохранения результата)
	PromptKey string // ключ в layout для текста запроса
	Validator func(string) bool
	ErrorKey  string // ключ в layout для сообщения об ошибке
}

var ErrCanceled = errors.New("input canceled")

// CollectInput собирает ответы пользователя пошагово.
// Возвращает map[name]значение или ошибку, если прервано.
func CollectInput(
	c tele.Context,
	inputManager *intele.InputManager,
	layout *layout.Layout,
	steps []InputStep,
) (map[string]string, error) {
	ic := collector.New()
	ic.Collect(c.Message()) // собираем начальное сообщение

	results := make(map[string]string, len(steps))
	isFirst := true

	for _, step := range steps {
		markup := layout.Markup(c, "habitsMenuBack") // кнопка "Назад" (можно параметризовать)

		// Отправляем запрос
		if isFirst {
			_ = c.Edit(layout.Text(c, step.PromptKey), markup)
		} else {
			_ = ic.Send(c, layout.Text(c, step.PromptKey), markup)
		}
		isFirst = false

		// Цикл ожидания корректного ввода
		for {
			response, err := inputManager.Get(context.Background(), c.Sender().ID, 0, nil)
			if response.Message != nil {
				ic.Collect(response.Message)
			}

			if response.Canceled {
				_ = ic.Clear(c, collector.ClearOptions{IgnoreErrors: true, ExcludeLast: true})
				return nil, ErrCanceled
			}

			if err != nil {
				log.Errorf("input error: %v", err)
				_ = ic.Send(c, layout.Text(c, "technical_issues"), markup)
				continue
			}

			// Если пришёл callback – считаем, что пользователь отменил
			if response.Callback != nil {
				_ = ic.Clear(c, collector.ClearOptions{IgnoreErrors: true})
				return nil, ErrCanceled
			}

			text := response.Message.Text
			if !step.Validator(text) {
				_ = ic.Send(c, layout.Text(c, step.ErrorKey), markup)
				continue
			}

			results[step.Name] = text
			_ = ic.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			break
		}
	}

	return results, nil
}
