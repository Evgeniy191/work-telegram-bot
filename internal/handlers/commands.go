package handlers

import (
	"fmt"
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards"
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

// HandleMyProjects показывает все проекты пользователя с кнопками управления
func HandleMyProjects(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	projects, err := database.GetUserProjects(chatID)

	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки проектов")
		bot.Send(msg)
		return
	}

	if len(projects) == 0 {
		msg := tgbotapi.NewMessage(chatID,
			"📋 У тебя пока нет проектов\n\n"+
				"Создай первый проект:\n"+
				"⚡ Быстрые действия → ➕ Новый проект")
		bot.Send(msg)
		return
	}

	// Отправляем каждый проект отдельным сообщением с кнопками
	for i, p := range projects {
		masterName := database.GetMasterNameByID(p.MasterID)

		// Получаем статистику задач и прогресс проекта
		total, pending, inProgress, completed, _ := database.CountProjectTasks(p.ID)
		progress := database.CalcProgress(total, completed)

		text := fmt.Sprintf(
			"📋 *Проект %d из %d*\n\n"+
				"*%s*\n"+
				"🔧 Тип: %s\n"+
				"💰 Бюджет: %.2f ₽\n"+
				"👷 Мастер: %s\n"+
				"📍 Статус: %s\n"+
				"📅 Создан: %s\n\n"+
				"📝 Задачи: %d (🕐 %d | ⚙️ %d | ✅ %d)\n"+
				"📊 Прогресс: %s %d%%",
			i+1,
			len(projects),
			p.Name,
			p.Type,
			p.Budget,
			masterName,
			p.Status,
			p.CreatedAt.Format("02.01.2006"),
			total, pending, inProgress, completed,
			progressBar(progress), progress,
		)

		// Создаём inline-кнопки
		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 Задачи", fmt.Sprintf("view_tasks_%d", p.ID)),
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить задачу", fmt.Sprintf("add_task_%d", p.ID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✏️ Изменить", fmt.Sprintf("edit_project_%d", p.ID)),
				tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("delete_project_%d", p.ID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)
	}
}
