package handlers

import (
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProcessUpdate обрабатывает одно обновление от Telegram:
// callback-кнопки, команду /cancel, FSM-диалоги и команды/reply-кнопки.
func ProcessUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) {
	// Inline-кнопки
	if update.CallbackQuery != nil {
		HandleCallback(bot, update, fsmManager)
		return
	}

	// Обрабатываем только сообщения
	if update.Message == nil {
		return
	}

	if update.Message.From != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}

	chatID := update.Message.Chat.ID

	// Отмена текущего действия
	if update.Message.Text == "/cancel" {
		if fsmManager.GetState(chatID) != fsm.StateIdle {
			fsmManager.ResetState(chatID)
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Действие отменено. Возврат в главное меню."))
		} else {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Нечего отменять. Ты в главном меню."))
		}
		return
	}

	// Сначала проверяем активный FSM-диалог
	if HandleFSMMessage(bot, update, fsmManager) {
		return
	}

	// Команды и reply-кнопки
	RouteCommand(bot, update, fsmManager)
}

// RouteCommand сопоставляет текст сообщения с обработчиком команды/кнопки меню.
func RouteCommand(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) {
	chatID := update.Message.Chat.ID

	switch update.Message.Text {
	case "/start":
		HandleStart(bot, update)
	case "/help", "❓ Помощь":
		HandleHelp(bot, update)
	case "/about":
		HandleAbout(bot, update)
	case "📋 Проекты", "/myprojects", "📋 Мои проекты":
		HandleMyProjects(bot, update)
	case "➕ Новый проект":
		if ensureAccess(bot, chatID, permManage) {
			StartProjectCreation(bot, chatID, fsmManager)
		}
	case "➕ Новая задача":
		if ensureAccess(bot, chatID, permEdit) {
			StartTaskCreation(bot, chatID, fsmManager)
		}
	case "📝 Задачи":
		HandleTasks(bot, update)
	case "👷 Мастера":
		HandleMasters(bot, update)
	case "📊 Отчёты":
		HandleReports(bot, update)
	case "⚙️ Настройки":
		HandleSettings(bot, update)
	case "🏠 Главное меню":
		HandleStart(bot, update)
	case "⚡ Быстрые действия":
		HandleQuickActions(bot, update)
	case "📝 Мои задачи":
		HandleMyTasks(bot, update)
	case "🌍 Язык":
		HandleLanguage(bot, update)
	case "🔔 Уведомления":
		HandleNotifications(bot, update)
	case "🎨 Тема":
		HandleTheme(bot, update)
	case "🔐 Безопасность":
		HandleSecurity(bot, update)
	case "📊 Формат отчётов":
		HandleReportFormat(bot, update)
	default:
		bot.Send(tgbotapi.NewMessage(chatID, "Я не понял эту команду. Используй меню ниже 👇"))
	}
}
