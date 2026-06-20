package main

import (
	"testing"
	"time"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestRunUpdates проверяет, что цикл обработки разбирает канал до закрытия.
// Пустое обновление не обращается к bot, поэтому передаём nil.
func TestRunUpdates(t *testing.T) {
	updates := make(chan tgbotapi.Update, 2)
	updates <- tgbotapi.Update{}            // нет Message и CallbackQuery → no-op
	updates <- tgbotapi.Update{UpdateID: 1} // тоже no-op
	close(updates)

	done := make(chan struct{})
	go func() {
		runUpdates(nil, updates, fsm.NewManager())
		close(done)
	}()

	select {
	case <-done:
		// канал разобран, функция вернулась
	case <-time.After(2 * time.Second):
		t.Fatal("runUpdates не завершилась после закрытия канала")
	}
}
