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
// markup — опциональная клавиатура (nil = без кнопок).
// firstStepSend — если true, первый шаг делает Send (новое сообщение), иначе Edit (редактирование существующего).
func CollectInput(
	c tele.Context,
	inputManager *intele.InputManager,
	layout *layout.Layout,
	steps []InputStep,
	markup *tele.ReplyMarkup,
	firstStepSend bool,
) (map[string]string, error) {
	ic := collector.New()
	if !firstStepSend {
		ic.Collect(c.Message())
	}

	results := make(map[string]string, len(steps))
	isFirst := true

	for _, step := range steps {
		var stepMarkup *tele.ReplyMarkup
		if markup != nil {
			stepMarkup = markup
		}

		if isFirst {
			if firstStepSend {
				var msg *tele.Message
				if stepMarkup != nil {
					msg, _ = c.Bot().Send(c.Chat(), layout.Text(c, step.PromptKey), stepMarkup)
				} else {
					msg, _ = c.Bot().Send(c.Chat(), layout.Text(c, step.PromptKey))
				}
				if msg != nil {
					ic.Collect(msg)
				}
			} else {
				if stepMarkup != nil {
					_ = c.Edit(layout.Text(c, step.PromptKey), stepMarkup)
				} else {
					_ = c.Edit(layout.Text(c, step.PromptKey))
				}
			}
		} else {
			if stepMarkup != nil {
				_ = ic.Send(c, layout.Text(c, step.PromptKey), stepMarkup)
			} else {
				_ = ic.Send(c, layout.Text(c, step.PromptKey))
			}
		}
		isFirst = false

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
				if stepMarkup != nil {
					_ = ic.Send(c, layout.Text(c, "technical_issues"), stepMarkup)
				} else {
					_ = ic.Send(c, layout.Text(c, "technical_issues"))
				}
				continue
			}

			if response.Callback != nil {
				_ = ic.Clear(c, collector.ClearOptions{IgnoreErrors: true})
				return nil, ErrCanceled
			}

			text := response.Message.Text
			if !step.Validator(text) {
				if stepMarkup != nil {
					_ = ic.Send(c, layout.Text(c, step.ErrorKey), stepMarkup)
				} else {
					_ = ic.Send(c, layout.Text(c, step.ErrorKey))
				}
				continue
			}

			results[step.Name] = text
			_ = ic.Clear(c, collector.ClearOptions{IgnoreErrors: true})
			break
		}
	}

	return results, nil
}
