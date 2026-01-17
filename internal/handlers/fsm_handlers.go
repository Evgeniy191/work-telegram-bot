package handlers

import (
	"fmt"
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
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
		// Пользователь вводит название проекта
		data := fsmManager.GetData(chatID)
		data.ProjectName = text
		fsmManager.SetData(chatID, data)

		// Переходим к следующему шагу: запрос бюджета
		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ Название проекта: <b>%s</b>\n\n💰 Теперь введи бюджет (в рублях):",
			text,
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true // Сообщение обработано FSM

	case fsm.StateCreatingProjectBudget:
		// Пользователь вводит бюджет
		data := fsmManager.GetData(chatID)
		data.ProjectBudget = text

		// Создаём проект
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"🎉 <b>Проект создан!</b>\n\n"+
				"📋 Название: %s\n"+
				"💰 Бюджет: %s ₽\n"+
				"📅 Дата: 17.01.2026\n\n"+
				"Проект добавлен в список!",
			data.ProjectName,
			data.ProjectBudget,
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)

		// Сбрасываем состояние
		fsmManager.ResetState(chatID)
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
	fsmManager.SetState(chatID, fsm.StateCreatingProject)

	msg := tgbotapi.NewMessage(chatID,
		"➕ <b>Создание нового проекта</b>\n\n"+
			"📝 Введи название проекта:")
	msg.ParseMode = "HTML"
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
