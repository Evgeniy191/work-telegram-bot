package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
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
	chatID := update.Message.Chat.ID

	masters, err := database.GetAllMasters()
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки списка мастеров")
		bot.Send(msg)
		return
	}

	if len(masters) == 0 {
		msg := tgbotapi.NewMessage(chatID, "👷 Список мастеров пуст.")
		msg.ReplyMarkup = keyboards.BackToMainMenu()
		bot.Send(msg)
		return
	}

	// Формируем текст со списком мастеров
	text := "👷 *Управление мастерами*\n\n" +
		"Список мастеров:\n"
	for i, m := range masters {
		specialty := m.Specialty
		if specialty == "" {
			specialty = "—"
		}
		text += fmt.Sprintf("%d. *%s* — %s\n", i+1, m.Name, specialty)
	}
	text += "\nЧтобы изменить ФИО, нажми кнопку ниже:"

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboards.MastersManageKeyboard()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// HandleReports обрабатывает кнопку "📊 Отчёты" — дашборд по портфелю проектов.
func HandleReports(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	dash, err := database.GetPortfolioDashboard(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка формирования отчёта")
		bot.Send(msg)
		return
	}

	if dash.Projects == 0 {
		msg := tgbotapi.NewMessage(chatID, "📊 Дашборд портфеля\n\nПока нет проектов для анализа.")
		msg.ReplyMarkup = keyboards.BackToMainMenu()
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, dashboardText(dash))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboards.BackToMainMenu()

	if _, err := bot.Send(msg); err != nil {
		log.Println("Ошибка отправки сообщения:", err)
	}
}

// dashboardText форматирует панель показателей портфеля для раздела «Отчёты».
func dashboardText(d database.PortfolioDashboard) string {
	overdueLine := "🟢 нет"
	if d.Overdue > 0 {
		overdueLine = fmt.Sprintf("🔴 %d", d.Overdue)
	}

	// Разбивка проектов по статусам — по строке на статус.
	var statusLines strings.Builder
	for _, sc := range d.StatusCounts {
		statusLines.WriteString(fmt.Sprintf("   • %s: %d\n", sc.Status, sc.Count))
	}

	// Освоение бюджета: сумма, процент и пометка перерасхода.
	spentLine := fmt.Sprintf("%.2f ₽ (%d%%)", d.Spent, d.BudgetPercent)
	if d.OverBudget {
		spentLine += " ⚠️ перерасход"
	}

	return fmt.Sprintf(
		"📊 *Дашборд портфеля*\n\n"+
			"📁 Проектов: %d\n"+
			"%s\n"+
			"💰 Бюджет: %.2f ₽\n"+
			"💸 Освоено: %s\n"+
			"💼 Остаток: %.2f ₽\n\n"+
			"📝 Задач всего: %d\n"+
			"✅ Выполнено: %d\n"+
			"⏰ Просрочено: %s\n"+
			"📈 Общий прогресс: %d%%",
		d.Projects,
		statusLines.String(),
		d.TotalBudget,
		spentLine,
		d.Remaining,
		d.Tasks,
		d.Completed,
		overdueLine,
		d.Progress,
	)
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

// HandleNotifications обрабатывает "🔔 Уведомления" — показывает состояние с переключателем.
func HandleNotifications(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	settings, err := database.GetUserSettings(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, "❌ Ошибка чтения настроек")
		bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(chatID, notificationsText(settings.Notifications))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboards.NotificationsKeyboard(settings.Notifications)

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

// HandleSecurity обрабатывает "🔐 Безопасность" — экран роли и доступа.
func HandleSecurity(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	role, err := database.GetUserRole(chatID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка чтения роли"))
		return
	}

	msg := tgbotapi.NewMessage(chatID, roleScreenText(role))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboards.RolesKeyboard(role)

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
