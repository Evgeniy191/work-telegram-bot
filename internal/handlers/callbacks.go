package handlers

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	"github.com/Evgeniy191/work-telegram-bot/internal/keyboards"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) {
	callback := update.CallbackQuery
	data := callback.Data
	chatID := callback.Message.Chat.ID

	log.Printf("📲 Получен callback: %s от пользователя %s", data, callback.From.UserName)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	bot.Request(callbackConfig)

	switch data {
	case "project_1":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 1: Монтаж труб")
		bot.Send(msg)

	case "project_2":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 2: Оборудование Б")
		bot.Send(msg)

	case "project_3":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 3: Монтаж м/к")
		bot.Send(msg)

	case "project_new":
		StartProjectCreation(bot, chatID, fsmManager)

	case "task_docs":
		msg := tgbotapi.NewMessage(chatID, "📄 Задача: Оформить документы")
		bot.Send(msg)

	case "task_materials":
		msg := tgbotapi.NewMessage(chatID, "📦 Задача: Заказать материалы")
		bot.Send(msg)

	case "task_report":
		msg := tgbotapi.NewMessage(chatID, "📊 Задача: Сдать отчёт")
		bot.Send(msg)

	case "back_projects":
		msg := tgbotapi.NewMessage(chatID, "📋 Выбери проект:")
		msg.ReplyMarkup = keyboards.ProjectsList()
		bot.Send(msg)

	case "confirm_yes":
		msg := tgbotapi.NewMessage(chatID, "✅ Действие подтверждено!")
		bot.Send(msg)

	case "confirm_no":
		msg := tgbotapi.NewMessage(chatID, "❌ Действие отменено.")
		bot.Send(msg)

	case "type_montazh", "type_remont", "type_ustanovka", "type_stroitelstvo":
		var typeName string
		var typeEmoji string

		switch data {
		case "type_montazh":
			typeName = "Монтаж"
			typeEmoji = "🔧"
		case "type_remont":
			typeName = "Ремонт"
			typeEmoji = "🛠️"
		case "type_ustanovka":
			typeName = "Установка"
			typeEmoji = "⚙️"
		case "type_stroitelstvo":
			typeName = "Строительство"
			typeEmoji = "🏗️"
		}

		userData := fsmManager.GetData(chatID)
		userData.ProjectType = typeName
		fsmManager.SetData(chatID, userData)

		fsmManager.SetState(chatID, fsm.StateCreatingProject)

		log.Printf("✅ CALLBACK: Установлено состояние=%s, тип=%s для chatID=%d",
			fsmManager.GetState(chatID), typeName, chatID)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Тип: <b>%s %s</b>\n\n"+
				"📝 Шаг 2/4: Введи название проекта:",
				typeEmoji, typeName))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_type")
		bot.Send(msg)

	case "back_to_type":
		fsmManager.SetState(chatID, fsm.StateCreatingProjectType)
		msg := tgbotapi.NewMessage(chatID,
			"◀️ Возврат назад\n\n"+
				"➕ <b>Создание нового проекта</b>\n\n"+
				"🔧 Шаг 1/4: Выбери тип проекта:")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.ProjectTypeKeyboard()
		bot.Send(msg)

	case "back_to_name":
		userData := fsmManager.GetData(chatID)
		fsmManager.SetState(chatID, fsm.StateCreatingProject)
		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("◀️ Возврат назад\n\n"+
				"✅ Тип: <b>%s %s</b>\n\n"+
				"📝 Шаг 2/4: Введи название проекта:",
				typeEmoji, userData.ProjectType))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_type")
		bot.Send(msg)

	case "back_to_budget":
		userData := fsmManager.GetData(chatID)
		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("◀️ Возврат назад\n\n"+
				"✅ Название: <b>%s</b>\n\n"+
				"💰 Шаг 3/4: Введи бюджет (в рублях):",
				userData.ProjectName))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_name")
		bot.Send(msg)

	case "master_ivanov", "master_petrov", "master_sidorov", "master_kuznetsov", "master_none":
		var masterName string
		switch data {
		case "master_ivanov":
			masterName = "Иванов Иван"
		case "master_petrov":
			masterName = "Петров Пётр"
		case "master_sidorov":
			masterName = "Сидоров Сергей"
		case "master_kuznetsov":
			masterName = "Кузнецов Андрей"
		case "master_none":
			masterName = "Не назначен"
		}

		userData := fsmManager.GetData(chatID)
		userData.ProjectMaster = masterName
		fsmManager.SetData(chatID, userData) // ✅ ДОБАВИЛ СОХРАНЕНИЕ!

		// СОХРАНЕНИЕ В БАЗУ ДАННЫХ
		budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)
		projectID, err := database.CreateProject(
			chatID,               // userID
			userData.ProjectType, // тип
			userData.ProjectName, // название
			budget,               // бюджет
			masterName,           // мастер
		)

		if err != nil {
			log.Printf("❌ Ошибка сохранения проекта: %v", err)
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка сохранения проекта в БД")
			bot.Send(msg)
			return
		}

		log.Printf("✅ Проект ID=%d сохранён в БД", projectID)

		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		budget, _ = strconv.ParseFloat(userData.ProjectBudget, 64)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"🎉 <b>Проект создан!</b>\n\n"+
				"%s <b>Тип:</b> %s\n"+
				"📋 <b>Название:</b> %s\n"+
				"💰 <b>Бюджет:</b> %s ₽\n"+
				"👷 <b>Ответственный:</b> %s\n"+
				"📅 <b>Дата создания:</b> %s\n"+
				"📍 <b>Статус:</b> В работе\n\n"+
				"✅ Проект добавлен в список!",
			typeEmoji,
			userData.ProjectType,
			userData.ProjectName,
			formatMoney(budget),
			masterName,
			time.Now().Format("02.01.2006"),
		))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", "edit_last_project"),
				tgbotapi.NewInlineKeyboardButtonData("◀️ В меню", "back_menu"),
			),
		)
		bot.Send(msg)
		// НЕ сбрасываем данные! Нужны для редактирования

	case "edit_last_project":
		userData := fsmManager.GetData(chatID)
		if userData.ProjectName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Нет данных проекта для редактирования.")
			bot.Send(msg)
			return
		}

		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✏️ <b>Редактирование проекта</b>\n\n"+
				"Текущие данные:\n"+
				"%s <b>Тип:</b> %s\n"+
				"📋 <b>Название:</b> %s\n"+
				"💰 <b>Бюджет:</b> %s ₽\n"+
				"👷 <b>Ответственный:</b> %s\n\n"+
				"Что изменить?",
			typeEmoji,
			userData.ProjectType,
			userData.ProjectName,
			formatMoney(budget),
			userData.ProjectMaster,
		))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔧 Тип", "edit_type"),
				tgbotapi.NewInlineKeyboardButtonData("📋 Название", "edit_name"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("💰 Бюджет", "edit_budget"),
				tgbotapi.NewInlineKeyboardButtonData("👷 Мастера", "edit_master"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Сохранить", "save_project"),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отменить", "back_menu"),
			),
		)
		bot.Send(msg)

	case "edit_type":
		msg := tgbotapi.NewMessage(chatID, "🔧 Выбери новый тип проекта:")
		msg.ReplyMarkup = keyboards.ProjectTypeKeyboard()
		bot.Send(msg)

	case "edit_name":
		fsmManager.SetState(chatID, fsm.StateCreatingProject)
		msg := tgbotapi.NewMessage(chatID, "📋 Введи новое название проекта:")
		bot.Send(msg)

	case "edit_budget":
		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID, "💰 Введи новый бюджет:")
		bot.Send(msg)

	case "edit_master":
		msg := tgbotapi.NewMessage(chatID, "👷 Выбери нового ответственного:")
		msg.ReplyMarkup = keyboards.MastersKeyboard()
		bot.Send(msg)

	case "save_project":
		userData := fsmManager.GetData(chatID)

		log.Printf("📝 SAVE: Тип=%s, Название=%s, Бюджет=%s, Мастер=%s",
			userData.ProjectType,
			userData.ProjectName,
			userData.ProjectBudget,
			userData.ProjectMaster)

		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ <b>Проект сохранён!</b>\n\n"+
				"%s <b>Тип:</b> %s\n"+
				"📋 <b>Название:</b> %s\n"+
				"💰 <b>Бюджет:</b> %s ₽\n"+
				"👷 <b>Ответственный:</b> %s\n"+
				"📅 <b>Обновлён:</b> %s",
			typeEmoji,
			userData.ProjectType,
			userData.ProjectName,
			formatMoney(budget),
			userData.ProjectMaster,
			time.Now().Format("02.01.2006 15:04"),
		))
		msg.ParseMode = "HTML"
		bot.Send(msg)
		fsmManager.ResetState(chatID)

	case "back_menu":
		msg := tgbotapi.NewMessage(chatID, "◀️ Возврат в главное меню")
		bot.Send(msg)
		fsmManager.ResetState(chatID)

	default:
		log.Printf("⚠️ Неизвестный callback_data: %s", data)
		msg := tgbotapi.NewMessage(chatID, "❌ Неизвестная команда")
		bot.Send(msg)
	}
}

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

func formatMoney(amount float64) string {
	intPart := int(amount)
	fracPart := amount - float64(intPart)

	intStr := strconv.Itoa(intPart)
	var result string

	for i, char := range reverse(intStr) {
		if i > 0 && i%3 == 0 {
			result = " " + result
		}
		result = string(char) + result
	}

	if fracPart > 0.009 {
		result += fmt.Sprintf(".%02d", int(fracPart*100+0.5))
	}

	return result
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
