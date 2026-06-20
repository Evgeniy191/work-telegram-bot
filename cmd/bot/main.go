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
	runUpdates(bot, updates, fsmManager)
}

// runUpdates читает обновления из канала и передаёт каждое в обработчик.
// Вынесено отдельно, чтобы цикл обработки можно было покрыть тестом.
func runUpdates(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, fsmManager *fsm.Manager) {
	for update := range updates {
		handlers.ProcessUpdate(bot, update, fsmManager)
	}
}
