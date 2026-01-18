package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards/inline"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleFSMMessage — обрабатывает сообщения в зависимости от состояния
func HandleFSMMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) bool {
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	state := fsmManager.GetState(chatID)

	log.Printf("🔍 FSM: Пользователь %d в состоянии '%s', сообщение: %s", chatID, state, text)

	switch state {
	case fsm.StateCreatingProject:
		data := fsmManager.GetData(chatID)
		data.ProjectName = text
		fsmManager.SetData(chatID, data)

		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ Название: <b>%s</b>\n\n"+
				"💰 Шаг 3/3: Введи бюджет (в рублях):",
			text,
		))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inline.BackButton("back_to_name")
		bot.Send(msg)
		return true

	case fsm.StateCreatingProjectBudget:
		// Пользователь вводит бюджет
		data := fsmManager.GetData(chatID)

		// Убираем пробелы по краям
		cleanText := strings.TrimSpace(text)

		// ВАЛИДАЦИЯ: Преобразуем в число (поддержка дробных)
		budget, err := strconv.ParseFloat(cleanText, 64)
		if err != nil {
			// text НЕ число
			msg := tgbotapi.NewMessage(chatID,
				"❌ <b>Ошибка!</b>\n\n"+
					"Бюджет должен быть <b>числом</b>.\n"+
					"Примеры: 500000 или 12500.50\n\n"+
					"💡 Попробуй ещё раз:")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return true
		}

		// ВАЛИДАЦИЯ: Проверяем положительность
		if budget < 0 {
			msg := tgbotapi.NewMessage(chatID,
				"❌ <b>Ошибка!</b>\n\n"+
					"Бюджет не может быть отрицательным.\n\n"+
					"💡 Попробуй ещё раз:")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return true
		}

		// Округляем до 2 знаков (копейки)
		budgetRounded := roundToTwoDecimals(budget)

		// Сохраняем как строку с форматированием
		// Сохраняем бюджет
		data.ProjectBudget = fmt.Sprintf("%.2f", budgetRounded)
		fsmManager.SetData(chatID, data)

		// ✅ ПЕРЕХОДИМ К ВЫБОРУ МАСТЕРА
		fsmManager.SetState(chatID, fsm.StateCreatingProjectMaster)

		msg := tgbotapi.NewMessage(chatID,
			"✅ Бюджет сохранён!\n\n"+
				"👷 Шаг 4/4: Назначь ответственного:")
		msg.ReplyMarkup = inline.MastersKeyboard()
		bot.Send(msg)
		return true

	case fsm.StateCreatingTask:
		// Пользователь вводит название задачи
		data := fsmManager.GetData(chatID)
		data.TaskName = text

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ <b>Задача создана!</b>\n\n"+
				"📝 Название: %s\n"+
				"📅 Срок: не установлен\n"+
				"👷 Исполнитель: не назначен",
			text,
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true

	case fsm.StateIdle:
		// Обычный режим — не обрабатываем здесь
		return false

	default:
		// Неизвестное состояние
		log.Printf("⚠️ FSM: Неизвестное состояние '%s'", state)
		return false
	}
}

// StartProjectCreation — начать создание проекта
func StartProjectCreation(bot *tgbotapi.BotAPI, chatID int64, fsmManager *fsm.Manager) {
	// Устанавливаем состояние "выбор типа"
	fsmManager.SetState(chatID, fsm.StateCreatingProjectType)

	msg := tgbotapi.NewMessage(chatID,
		"➕ <b>Создание нового проекта</b>\n\n"+
			"🔧 Шаг 1/3: Выбери тип проекта:")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = inline.ProjectTypeKeyboard() // ← Inline-кнопки!
	bot.Send(msg)
}

// StartTaskCreation — начать создание задачи
func StartTaskCreation(bot *tgbotapi.BotAPI, chatID int64, fsmManager *fsm.Manager) {
	fsmManager.SetState(chatID, fsm.StateCreatingTask)

	msg := tgbotapi.NewMessage(chatID,
		"➕ <b>Создание новой задачи</b>\n\n"+
			"📝 Введи название задачи:")
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

// roundToTwoDecimals — округляет до 2 знаков после запятой
func roundToTwoDecimals(num float64) float64 {
	return float64(int(num*100+0.5)) / 100
}
