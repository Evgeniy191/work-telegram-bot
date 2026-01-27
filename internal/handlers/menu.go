package handlers

import (
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleProjects обрабатывает кнопку "📋 Проекты"
func HandleProjects(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"📋 Управление проектами\n\n"+
			"В этом разделе ты сможешь:\n"+
			"• Просматривать список проектов\n"+
			"• Создавать новые проекты\n"+
			"• Редактировать существующие\n"+
			"• Отслеживать статус выполнения\n\n"+
			"🔜 Функционал в разработке...",
	)

	// Добавляем кнопку возврата
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleTasks обрабатывает кнопку "📝 Задачи"
func HandleTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"📝 Управление задачами\n\n"+
			"В этом разделе ты сможешь:\n"+
			"• Создавать новые задачи\n"+
			"• Назначать задачи мастерам\n"+
			"• Отслеживать выполнение\n"+
			"• Устанавливать дедлайны\n\n"+
			"🔜 Функционал в разработке...",
	)

	// Добавляем кнопку возврата
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleMasters обрабатывает кнопку "👷 Мастера"
func HandleMasters(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"👷 Управление мастерами\n\n"+
			"В этом разделе ты сможешь:\n"+
			"• Просматривать список мастеров\n"+
			"• Добавлять новых мастеров\n"+
			"• Просматривать их задачи\n"+
			"• Отслеживать загруженность\n\n"+
			"🔜 Функционал в разработке...",
	)

	// Добавляем кнопку возврата
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleReports обрабатывает кнопку "📊 Отчёты"
func HandleReports(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"📊 Отчёты\n\n"+
			"В этом разделе ты сможешь:\n"+
			"• Формировать отчёты по проектам\n"+
			"• Просматривать статистику\n"+
			"• Экспортировать данные\n"+
			"• Анализировать эффективность\n\n"+
			"🔜 Функционал в разработке...",
	)

	// Добавляем кнопку возврата
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleSettings обрабатывает кнопку "⚙️ Настройки"
func HandleSettings(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	HandleSettingsMenu(bot, update) // ← УБРАЛ handlers.
}

// HandleQuickActions показывает меню быстрых действий
func HandleQuickActions(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"⚡ Быстрые действия\n\n"+
			"Выбери нужное действие:",
	)

	msg.ReplyMarkup = keyboards.QuickActionsMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleNewProject обрабатывает "➕ Новый проект"
func HandleNewProject(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"➕ Создание нового проекта\n\n"+
			"📝 Введи название проекта:",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleNewTask обрабатывает "➕ Новая задача"
func HandleNewTask(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"➕ Создание новой задачи\n\n"+
			"📝 Введи описание задачи:",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleMyTasks обрабатывает "📝 Мои задачи"
func HandleMyTasks(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"📝 Мои задачи\n\n"+
			"🔄 Загрузка...\n\n"+
			"(Список задач пока пуст)\n\n"+
			"🔜 Будут задачи из БД",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleSettingsMenu показывает меню настроек
func HandleSettingsMenu(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"⚙️ Настройки\n\n"+
			"Выбери раздел настроек:",
	)

	msg.ReplyMarkup = keyboards.SettingsMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleLanguage обрабатывает "🌍 Язык"
func HandleLanguage(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🌍 Язык интерфейса\n\n"+
			"Текущий: Русский 🇷🇺\n\n"+
			"Доступно:\n"+
			"🇷🇺 Русский\n"+
			"🇺🇸 English\n\n"+
			"🔜 Выбор языка",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleNotifications обрабатывает "🔔 Уведомления"
func HandleNotifications(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🔔 Уведомления\n\n"+
			"Текущие настройки:\n"+
			"✅ Уведомления о задачах\n"+
			"✅ Уведомления о дедлайнах\n"+
			"❌ Отчёты\n\n"+
			"🔜 Настройка уведомлений",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleTheme обрабатывает "🎨 Тема"
func HandleTheme(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🎨 Тема интерфейса\n\n"+
			"Текущая: Светлая ☀️\n\n"+
			"Доступно:\n"+
			"☀️ Светлая\n"+
			"🌙 Тёмная\n"+
			"🌈 Авто\n\n"+
			"🔜 Выбор темы",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleSecurity обрабатывает "🔐 Безопасность"
func HandleSecurity(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"🔐 Безопасность\n\n"+
			"Безопасность аккаунта:\n"+
			"✅ Двухфакторная аутентификация\n"+
			"✅ Автоматический выход\n\n"+
			"🔜 Настройки безопасности",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleReportFormat обрабатывает "📊 Формат отчётов"
func HandleReportFormat(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"📊 Формат отчётов\n\n"+
			"Текущий: PDF 📄\n\n"+
			"Доступно:\n"+
			"📄 PDF\n"+
			"📊 Excel\n"+
			"📈 Графики\n\n"+
			"🔜 Выбор формата",
	)
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}
