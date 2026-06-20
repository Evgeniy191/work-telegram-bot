package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards"
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
		msg.ReplyMarkup = keyboards.BackButton("back_to_name")
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
		// Сохраняем бюджет
		data.ProjectBudget = fmt.Sprintf("%.2f", budgetRounded)
		fsmManager.SetData(chatID, data)

		// ✅ ПЕРЕХОДИМ К ВЫБОРУ МАСТЕРА
		fsmManager.SetState(chatID, fsm.StateCreatingProjectMaster)

		msg := tgbotapi.NewMessage(chatID,
			"✅ Бюджет сохранён!\n\n"+
				"👷 Шаг 4/4: Назначь ответственного:")
		msg.ReplyMarkup = keyboards.MastersKeyboard()
		bot.Send(msg)
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

		// Редактирование названия проекта
	case fsm.StateEditingProjectName:
		data := fsmManager.GetData(chatID)
		newName := strings.TrimSpace(text)

		if newName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Название не может быть пустым. Попробуй ещё раз:")
			bot.Send(msg)
			return true
		}

		// Получаем текущий проект
		project, err := database.GetProjectByID(data.EditingProjectID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Проект не найден")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		// Обновляем проект
		masterName := database.GetMasterNameByID(project.MasterID)
		err = database.UpdateProject(data.EditingProjectID, newName, project.Budget, masterName)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления проекта")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Название обновлено*\n\n"+
				"Новое название: *%s*", newName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true

	// Редактирование бюджета проекта
	case fsm.StateEditingProjectBudget:
		data := fsmManager.GetData(chatID)
		cleanText := strings.TrimSpace(text)

		// Валидация числа
		budget, err := strconv.ParseFloat(cleanText, 64)
		if err != nil || budget < 0 {
			msg := tgbotapi.NewMessage(chatID,
				"❌ Ошибка!\n\n"+
					"Бюджет должен быть положительным числом.\n"+
					"Примеры: 500000 или 12500.50\n\n"+
					"💡 Попробуй ещё раз:")
			bot.Send(msg)
			return true
		}

		// Получаем текущий проект
		project, err := database.GetProjectByID(data.EditingProjectID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Проект не найден")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		// Обновляем проект
		masterName := database.GetMasterNameByID(project.MasterID)
		err = database.UpdateProject(data.EditingProjectID, project.Name, budget, masterName)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления проекта")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Бюджет обновлён*\n\n"+
				"Новый бюджет: *%.2f ₽*", budget))
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true
	case fsm.StateCreatingTaskName:
		data := fsmManager.GetData(chatID)
		taskName := strings.TrimSpace(text)

		if taskName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Название не может быть пустым")
			bot.Send(msg)
			return true
		}

		data.TaskName = taskName
		fsmManager.SetData(chatID, data)
		fsmManager.SetState(chatID, fsm.StateCreatingTaskDescription)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("📝 *Название:* %s\n\n"+
				"📄 Введи описание задачи (или /skip для пропуска):", taskName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return true

	case fsm.StateCreatingTaskDescription:
		data := fsmManager.GetData(chatID)
		taskDesc := strings.TrimSpace(text)

		// Пропуск описания
		if text == "/skip" {
			taskDesc = ""
		}

		data.TaskDescription = taskDesc
		fsmManager.SetData(chatID, data)
		fsmManager.SetState(chatID, fsm.StateCreatingTaskDeadline)

		msg := tgbotapi.NewMessage(chatID,
			"📅 Введи срок задачи в формате *ДД.ММ.ГГГГ*\n"+
				"(или /skip — без срока):")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		return true

	case fsm.StateCreatingTaskDeadline:
		data := fsmManager.GetData(chatID)

		var deadline *time.Time
		if text != "/skip" {
			d, err := database.ParseDeadline(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID,
					"❌ Неверный формат даты.\n"+
						"Пример: 25.12.2026\n\n"+
						"💡 Попробуй ещё раз (или /skip):")
				bot.Send(msg)
				return true
			}
			deadline = d
		}

		taskID, err := database.CreateTask(data.TaskProjectID, data.TaskName, data.TaskDescription, deadline)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка создания задачи")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		deadlineText := "не указан"
		if deadline != nil {
			deadlineText = deadline.Format("02.01.2006")
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Задача создана!*\n\n"+
				"📝 %s\n"+
				"📄 %s\n"+
				"📅 Срок: %s\n"+
				"🕐 Статус: Ожидает\n\n"+
				"ID: %d", data.TaskName, data.TaskDescription, deadlineText, taskID))
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true

	// Редактирование названия задачи
	case fsm.StateEditingTaskName:
		data := fsmManager.GetData(chatID)
		newName := strings.TrimSpace(text)
		if newName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Название не может быть пустым. Попробуй ещё раз:")
			bot.Send(msg)
			return true
		}

		if err := database.UpdateTaskName(data.EditingTaskID, newName); err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления задачи")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Название задачи обновлено*\n\nНовое название: *%s*", newName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		fsmManager.ResetState(chatID)
		return true

	// Редактирование описания задачи
	case fsm.StateEditingTaskDescription:
		data := fsmManager.GetData(chatID)
		newDesc := strings.TrimSpace(text)
		if text == "/clear" {
			newDesc = ""
		}

		if err := database.UpdateTaskDescription(data.EditingTaskID, newDesc); err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления задачи")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		msg := tgbotapi.NewMessage(chatID, "✅ *Описание задачи обновлено*")
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		fsmManager.ResetState(chatID)
		return true

	// Редактирование срока задачи
	case fsm.StateEditingTaskDeadline:
		data := fsmManager.GetData(chatID)

		var deadline *time.Time
		if text != "/clear" {
			d, err := database.ParseDeadline(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID,
					"❌ Неверный формат даты.\n"+
						"Пример: 25.12.2026\n\n"+
						"💡 Попробуй ещё раз (или /clear — убрать срок):")
				bot.Send(msg)
				return true
			}
			deadline = d
		}

		if err := database.UpdateTaskDeadline(data.EditingTaskID, deadline); err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка обновления срока")
			bot.Send(msg)
			fsmManager.ResetState(chatID)
			return true
		}

		deadlineText := "убран"
		if deadline != nil {
			deadlineText = deadline.Format("02.01.2006")
		}
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *Срок задачи обновлён*\n\nНовый срок: *%s*", deadlineText))
		msg.ParseMode = "Markdown"
		bot.Send(msg)
		fsmManager.ResetState(chatID)
		return true

	// Редактирование ФИО мастера
	case fsm.StateEditingMasterName:
		data := fsmManager.GetData(chatID)
		newName := strings.TrimSpace(text)

		if newName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ ФИО не может быть пустым. Попробуй ещё раз:")
			bot.Send(msg)
			return true
		}

		if err := database.UpdateMasterName(data.EditingMasterID, newName); err != nil {
			msg := tgbotapi.NewMessage(chatID,
				fmt.Sprintf("❌ Не удалось обновить ФИО.\n%s\n\n💡 Попробуй другое ФИО:", err.Error()))
			bot.Send(msg)
			return true
		}

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ *ФИО обновлено*\n\nНовое ФИО: *%s*", newName))
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		fsmManager.ResetState(chatID)
		return true

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
	msg.ReplyMarkup = keyboards.ProjectTypeKeyboard() // ← Inline-кнопки!
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

// roundToTwoDecimals — округляет до 2 знаков после запятой
func roundToTwoDecimals(num float64) float64 {
	return float64(int(num*100+0.5)) / 100
}
