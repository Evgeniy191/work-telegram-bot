package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/Evgeniy191/work-telegram-bot/internal/config"
	"github.com/Evgeniy191/work-telegram-bot/internal/handlers"
)

func main() {
	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	// Загружаем конфигурацию
	cfg := config.Load()

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = cfg.Debug
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Обрабатываем обновления
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		// Обработка команд
		switch update.Message.Text {
		case "/start":
			handlers.HandleStart(bot, update)
		case "/help", "❓ Помощь":
			handlers.HandleHelp(bot, update)
		case "/about":
			handlers.HandleAbout(bot, update)
		case "📋 Проекты":
			handlers.HandleProjects(bot, update)
		case "📝 Задачи":
			handlers.HandleTasks(bot, update)
		case "👷 Мастера":
			handlers.HandleMasters(bot, update)
		case "📊 Отчёты":
			handlers.HandleReports(bot, update)
		case "⚙️ Настройки":
			handlers.HandleSettings(bot, update)
		case "🏠 Главное меню":
			handlers.HandleStart(bot, update)
			// В блоке switch добавь:
		case "⚡ Быстрые действия":
			handlers.HandleQuickActions(bot, update)
		case "➕ Новый проект":
			handlers.HandleNewProject(bot, update)
		case "➕ Новая задача":
			handlers.HandleNewTask(bot, update)
		case "📋 Мои проекты":
			handlers.HandleMyProjects(bot, update)
		case "📝 Мои задачи":
			handlers.HandleMyTasks(bot, update)
		case "🌍 Язык":
			handlers.HandleLanguage(bot, update)
		case "🔔 Уведомления":
			handlers.HandleNotifications(bot, update)
		case "🎨 Тема":
			handlers.HandleTheme(bot, update)
		case "🔐 Безопасность":
			handlers.HandleSecurity(bot, update)
		case "📊 Формат отчётов":
			handlers.HandleReportFormat(bot, update)
		default:
			// Обработка обычных сообщений
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не понял эту команду. Используй меню ниже 👇")
			bot.Send(msg)
		}
	}
}
