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
		case "/help":
			handlers.HandleHelp(bot, update)
		case "/about":
			handlers.HandleAbout(bot, update)
		default:
			// Обработка обычных сообщений (пока просто эхо)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ты написал: "+update.Message.Text)
			bot.Send(msg)
		}
	}
}
