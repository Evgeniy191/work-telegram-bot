package main

import (
	"log"
	"os"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем переменные из .env
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем системные переменные")
	}

	// 🗄️ ИНИЦИАЛИЗАЦИЯ БД (ДОБАВЬ ЭТО)
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Ошибка инициализации БД: %v", err)
	}
	defer database.CloseDB()

	token := os.Getenv("TELEGRAM_TOKEN") // Вместо "BOT_TOKEN"
	if token == "" {
		log.Fatal("❌ TELEGRAM_TOKEN не установлен в .env")
	}

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true // ✅ Вместо cfg.Debug
	log.Printf("Авторизован как %s", bot.Self.UserName)

	// Создаём FSM менеджер
	fsmManager := fsm.NewManager()

	log.Println("🔄 FSM менеджер инициализирован")

	// Настраиваем получение обновлений
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	// Обрабатываем обновления
	for update := range updates {
		// Inline-кнопки
		if update.CallbackQuery != nil {
			handlers.HandleCallback(bot, update, fsmManager)
			continue
		}

		// Сообщения
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Text == "/cancel" {
			chatID := update.Message.Chat.ID
			currentState := fsmManager.GetState(chatID)

			if currentState != fsm.StateIdle {
				fsmManager.ResetState(chatID)
				msg := tgbotapi.NewMessage(chatID, "❌ Действие отменено. Возврат в главное меню.")
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "⚠️ Нечего отменять. Ты в главном меню.")
				bot.Send(msg)
			}
			continue // Прерываем цикл
		}

		// Проверяем FSM ПЕРЕД switch
		if handlers.HandleFSMMessage(bot, update, fsmManager) {
			continue // Сообщение обработано FSM
		}

		// Обработка команд и Reply-кнопок
		switch update.Message.Text {
		case "/start":
			handlers.HandleStart(bot, update)
		case "/help", "❓ Помощь":
			handlers.HandleHelp(bot, update)
		case "/about":
			handlers.HandleAbout(bot, update)
		case "📋 Проекты":
			handlers.HandleMyProjects(bot, update)
		case "/myprojects":
			handlers.HandleMyProjects(bot, update)
		case "➕ Новый проект":
			// Запускаем FSM диалог
			handlers.StartProjectCreation(bot, update.Message.Chat.ID, fsmManager)
		case "➕ Новая задача":
			// Запускаем FSM диалог
			handlers.StartTaskCreation(bot, update.Message.Chat.ID, fsmManager)
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
