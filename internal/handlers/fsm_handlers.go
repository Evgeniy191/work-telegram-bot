package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards/inline"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandleFSMMessage — обрабатывает сообщения в зависимости от состояния
func HandleFSMMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) bool {
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	state := fsmManager.GetState(chatID)

	log.Printf("🔍 FSM: Пользователь %d в состоянии '%s', сообщение: %s", chatID, state, text)

	switch state {
	case fsm.StateCreatingProject:
		data := fsmManager.GetData(chatID)
		data.ProjectName = text
		fsmManager.SetData(chatID, data)

		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ Название: <b>%s</b>\n\n"+
				"💰 Шаг 3/3: Введи бюджет (в рублях):",
			text,
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		return true

	case fsm.StateCreatingProjectBudget:
		// Пользователь вводит бюджет
		data := fsmManager.GetData(chatID)

		// Убираем пробелы по краям
		cleanText := strings.TrimSpace(text)

		// ВАЛИДАЦИЯ: Преобразуем в число (поддержка дробных)
		budget, err := strconv.ParseFloat(cleanText, 64)
		if err != nil {
			// text НЕ число
			msg := tgbotapi.NewMessage(chatID,
				"❌ <b>Ошибка!</b>\n\n"+
					"Бюджет должен быть <b>числом</b>.\n"+
					"Примеры: 500000 или 12500.50\n\n"+
					"💡 Попробуй ещё раз:")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return true
		}

		// ВАЛИДАЦИЯ: Проверяем положительность
		if budget < 0 {
			msg := tgbotapi.NewMessage(chatID,
				"❌ <b>Ошибка!</b>\n\n"+
					"Бюджет не может быть отрицательным.\n\n"+
					"💡 Попробуй ещё раз:")
			msg.ParseMode = "HTML"
			bot.Send(msg)
			return true
		}

		// Округляем до 2 знаков (копейки)
		budgetRounded := roundToTwoDecimals(budget)

		// Сохраняем как строку с форматированием
		data.ProjectBudget = fmt.Sprintf("%.2f", budgetRounded)

		// Создаём проект
		// Определяем эмодзи типа
		typeEmoji := getProjectTypeEmoji(data.ProjectType)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"🎉 <b>Проект создан!</b>\n\n"+
				"%s <b>Тип:</b> %s\n"+
				"📋 <b>Название:</b> %s\n"+
				"💰 <b>Бюджет:</b> %s ₽\n"+
				"📅 <b>Дата создания:</b> %s\n"+
				"📍 <b>Статус:</b> В работе\n\n"+
				"✅ Проект добавлен в список!",
			typeEmoji,
			data.ProjectType,
			data.ProjectName,
			formatMoney(budgetRounded),
			time.Now().Format("02.01.2006"),
		))

		msg.ParseMode = "HTML"
		bot.Send(msg)

		// Сбрасываем состояние
		fsmManager.ResetState(chatID)
		return true

	case fsm.StateCreatingTask:
		// Пользователь вводит название задачи
		data := fsmManager.GetData(chatID)
		data.TaskName = text

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ <b>Задача создана!</b>\n\n"+
				"📝 Название: %s\n"+
				"📅 Срок: не установлен\n"+
				"👷 Исполнитель: не назначен",
			text,
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true

	case fsm.StateIdle:
		// Обычный режим — не обрабатываем здесь
		return false

	default:
		// Неизвестное состояние
		log.Printf("⚠️ FSM: Неизвестное состояние '%s'", state)
		return false
	}
}

// StartProjectCreation — начать создание проекта
func StartProjectCreation(bot *tgbotapi.BotAPI, chatID int64, fsmManager *fsm.Manager) {
	// Устанавливаем состояние "выбор типа"
	fsmManager.SetState(chatID, fsm.StateCreatingProjectType)

	msg := tgbotapi.NewMessage(chatID,
		"➕ <b>Создание нового проекта</b>\n\n"+
			"🔧 Шаг 1/3: Выбери тип проекта:")
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = inline.ProjectTypeKeyboard() // ← Inline-кнопки!
	bot.Send(msg)
}

// StartTaskCreation — начать создание задачи
func StartTaskCreation(bot *tgbotapi.BotAPI, chatID int64, fsmManager *fsm.Manager) {
	fsmManager.SetState(chatID, fsm.StateCreatingTask)

	msg := tgbotapi.NewMessage(chatID,
		"➕ <b>Создание новой задачи</b>\n\n"+
			"📝 Введи название задачи:")
	msg.ParseMode = "HTML"
	bot.Send(msg)
}

// formatMoney — форматирует число с разделителями тысяч
// 500000.50 → "500 000.50"
func formatMoney(amount float64) string {
	// Разбиваем на целую и дробную часть
	intPart := int(amount)
	fracPart := amount - float64(intPart)

	// Форматируем целую часть с разделителями
	intStr := strconv.Itoa(intPart)
	var result string

	// Добавляем пробелы каждые 3 цифры справа
	for i, char := range reverse(intStr) {
		if i > 0 && i%3 == 0 {
			result = " " + result
		}
		result = string(char) + result
	}

	// Добавляем дробную часть если есть
	if fracPart > 0.009 { // Учитываем погрешность float
		result += fmt.Sprintf(".%02d", int(fracPart*100+0.5))
	}

	return result
}

// reverse — переворачивает строку
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// roundToTwoDecimals — округляет до 2 знаков после запятой
func roundToTwoDecimals(num float64) float64 {
	return float64(int(num*100+0.5)) / 100
}

// getProjectTypeEmoji — возвращает эмодзи для типа проекта
func getProjectTypeEmoji(projectType string) string {
	switch projectType {
	case "Монтаж":
		return "🔧"
	case "Ремонт":
		return "🛠️"
	case "Установка":
		return "⚙️"
	case "Строительство":
		return "🏗️"
	default:
		return "📋"
	}
}
