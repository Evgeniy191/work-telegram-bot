package handlers

import (
	"fmt"
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards/inline"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleCallback — обрабатывает все callback queries от inline-кнопок
func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) {
	callback := update.CallbackQuery
	data := callback.Data
	chatID := callback.Message.Chat.ID

	log.Printf("📲 Получен callback: %s от пользователя %s", data, callback.From.UserName)

	// Подтверждаем получение (убирает спиннер)
	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	if _, err := bot.Request(callbackConfig); err != nil {
		log.Printf("❌ Ошибка AnswerCallbackQuery: %v", err)
	}

	// Обработка по callback_data
	switch data {
	case "project_1":
		msg := tgbotapi.NewMessage(chatID, "🏗️ <b>Проект: Монтаж труб А</b>\n\n"+
			"📍 Статус: В работе ✅\n"+
			"👷 Бригада: 5 человек\n"+
			"📅 Срок: 15.02.2026\n\n"+
			"Выбери задачу:")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inline.TasksList()
		bot.Send(msg)

	case "project_2":
		msg := tgbotapi.NewMessage(chatID, "🔧 <b>Проект: Оборудование Б</b>\n\n"+
			"📍 Статус: Завершён ✅\n"+
			"📅 Дата завершения: 10.01.2026")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case "project_3":
		msg := tgbotapi.NewMessage(chatID, "🔩 <b>Проект: Монтаж м/к</b>\n\n"+
			"📍 Статус: Планирование 📝\n"+
			"📅 Старт: 01.03.2026")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case "project_new":
		// ✅ ЗАПУСКАЕМ FSM!
		StartProjectCreation(bot, chatID, fsmManager)

	case "task_docs":
		msg := tgbotapi.NewMessage(chatID, "✅ <b>Задача: Проверить документы</b>\n\n"+
			"Документы проверены! 🎉\n"+
			"Отметить выполненной?")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = inline.ConfirmAction()
		bot.Send(msg)

	case "task_materials":
		msg := tgbotapi.NewMessage(chatID, "⏳ <b>Задача: Закупка материалов</b>\n\n"+
			"Статус: В процессе\n"+
			"💰 Бюджет: 150 000 ₽")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case "task_report":
		msg := tgbotapi.NewMessage(chatID, "📊 <b>Отчёт по проекту</b>\n\n"+
			"✅ Выполнено: 65%\n"+
			"⏳ В процессе: 25%\n"+
			"❌ Не начато: 10%")
		msg.ParseMode = "HTML"
		bot.Send(msg)

	case "back_projects":
		msg := tgbotapi.NewMessage(chatID, "◀️ Возврат к списку проектов:")
		msg.ReplyMarkup = inline.ProjectsList()
		bot.Send(msg)

	case "confirm_yes":
		msg := tgbotapi.NewMessage(chatID, "✅ Действие подтверждено!\n\nЗадача отмечена как выполненная.")
		bot.Send(msg)

	case "confirm_no":
		msg := tgbotapi.NewMessage(chatID, "❌ Действие отменено.")
		bot.Send(msg)

	case "type_montazh", "type_remont", "type_ustanovka", "type_stroitelstvo":
		// Определяем название типа по callback_data
		var typeName string
		var typeEmoji string

		switch data {
		case "type_montazh":
			typeName = "Монтаж"
			typeEmoji = "🔧"
		case "type_remont":
			typeName = "Ремонт"
			typeEmoji = "🛠️"
		case "type_ustanovka":
			typeName = "Установка"
			typeEmoji = "⚙️"
		case "type_stroitelstvo":
			typeName = "Строительство"
			typeEmoji = "🏗️"
		}

		// Сохраняем тип в данных пользователя
		userData := fsmManager.GetData(chatID)
		userData.ProjectType = typeName
		fsmManager.SetData(chatID, userData)

		// Переходим к вводу названия
		fsmManager.SetState(chatID, fsm.StateCreatingProject)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Тип: <b>%s %s</b>\n\n"+
				"📝 Шаг 2/3: Введи название проекта:",
				typeEmoji, typeName))
		msg.ParseMode = "HTML"
		bot.Send(msg)

	default:
		log.Printf("⚠️ Неизвестный callback_data: %s", data)
		msg := tgbotapi.NewMessage(chatID, "❌ Неизвестная команда")
		bot.Send(msg)
	}
}
