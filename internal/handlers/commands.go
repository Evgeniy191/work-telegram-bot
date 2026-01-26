package handlers

import (
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleStart обрабатывает команду /start
func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"👋 Добро пожаловать! Я твой рабочий помощник.\n\n"+
			"Что я умею:\n"+
			"📋 Управление проектами\n"+
			"📝 Создание задач\n"+
			"👷 Управление мастерами\n"+
			"📊 Формирование отчётов\n\n"+
			"Используй меню ниже для навигации. 👇",
	)

	// Добавляем главное меню
	msg.ReplyMarkup = keyboards.MainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения: ", err)
	}
}

// HandleHelp обрабатывает команду /help
func HandleHelp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"❓ Помощь\n\n"+
			"Доступные команды:\n"+
			"/start - Начать работу\n"+
			"/help - Показать эту справку\n\n"+
			"Используй кнопки меню для быстрого доступа к функциям:\n"+
			"📋 Проекты - Управление проектами\n"+
			"📝 Задачи - Управление задачами\n"+
			"👷 Мастера - Управление мастерами\n"+
			"📊 Отчёты - Просмотр отчётов\n"+
			"⚙️ Настройки - Настройки бота\n",
	)

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения", err)
	}
}

// HandleAbout показывает информацию о боте
func HandleAbout(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"ℹ️ О боте\n\n"+
			"Версия: 1.0.0\n"+
			"Разработчик: Evgeniy191\n"+
			"Назначение: Управление проектами\n\n"+
			"Функции:\n"+
			"• 📋 Управление проектами\n"+
			"• 📝 Создание и назначение задач\n"+
			"• 👷 Управление мастерами\n"+
			"• 📊 Формирование отчётов\n"+
			"• 📅 Отслеживание графика работ\n"+
			"• 📎 Управление документацией\n\n"+
			"Статус: ✅ В разработке\n"+
			"GitHub: github.com/Evgeniy191/work-telegram-bot",
	)

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка тправки сообщения: ", err)
	}
}
