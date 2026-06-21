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

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, fsmManager *fsm.Manager) {
	callback := update.CallbackQuery
	data := callback.Data
	chatID := callback.Message.Chat.ID

	log.Printf("📲 Получен callback: %s от пользователя %s", data, callback.From.UserName)

	callbackConfig := tgbotapi.NewCallback(callback.ID, "")
	bot.Request(callbackConfig)

	// Контроль доступа по роли пользователя
	if !checkCallbackAccess(bot, chatID, data) {
		return
	}

	switch {
	case data == "project_1":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 1: Монтаж труб")
		bot.Send(msg)

	case data == "project_2":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 2: Оборудование Б")
		bot.Send(msg)

	case data == "project_3":
		msg := tgbotapi.NewMessage(chatID, "📋 Проект 3: Монтаж м/к")
		bot.Send(msg)

	case data == "project_new":
		StartProjectCreation(bot, chatID, fsmManager)

	case data == "task_docs":
		msg := tgbotapi.NewMessage(chatID, "📄 Задача: Оформить документы")
		bot.Send(msg)

	case data == "task_materials":
		msg := tgbotapi.NewMessage(chatID, "📦 Задача: Заказать материалы")
		bot.Send(msg)

	case data == "task_report":
		msg := tgbotapi.NewMessage(chatID, "📊 Задача: Сдать отчёт")
		bot.Send(msg)

	case data == "back_projects":
		msg := tgbotapi.NewMessage(chatID, "📋 Выбери проект:")
		msg.ReplyMarkup = keyboards.ProjectsList()
		bot.Send(msg)

	case data == "confirm_yes":
		msg := tgbotapi.NewMessage(chatID, "✅ Действие подтверждено!")
		bot.Send(msg)

	case data == "confirm_no":
		msg := tgbotapi.NewMessage(chatID, "❌ Действие отменено.")
		bot.Send(msg)

	// ИСПРАВЛЕНО: используем || вместо запятых
	case data == "type_montazh" || data == "type_remont" || data == "type_ustanovka" || data == "type_stroitelstvo":
		var typeName string
		var typeEmoji string

		switch {
		case data == "type_montazh":
			typeName = "Монтаж"
			typeEmoji = "🔧"
		case data == "type_remont":
			typeName = "Ремонт"
			typeEmoji = "🛠️"
		case data == "type_ustanovka":
			typeName = "Установка"
			typeEmoji = "⚙️"
		case data == "type_stroitelstvo":
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
			fmt.Sprintf("✅ Тип: %s %s\n\n"+
				"📝 Шаг 2/4: Введи название проекта:",
				typeEmoji, typeName))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_type")
		bot.Send(msg)

	case data == "back_to_type":
		fsmManager.SetState(chatID, fsm.StateCreatingProjectType)
		msg := tgbotapi.NewMessage(chatID,
			"◀️ Возврат назад\n\n"+
				"➕ Создание нового проекта\n\n"+
				"🔧 Шаг 1/4: Выбери тип проекта:")
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.ProjectTypeKeyboard()
		bot.Send(msg)

	case data == "back_to_name":
		userData := fsmManager.GetData(chatID)
		fsmManager.SetState(chatID, fsm.StateCreatingProject)
		typeEmoji := getProjectTypeEmoji(userData.ProjectType)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("◀️ Возврат назад\n\n"+
				"✅ Тип: %s %s\n\n"+
				"📝 Шаг 2/4: Введи название проекта:",
				typeEmoji, userData.ProjectType))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_type")
		bot.Send(msg)

	case data == "back_to_budget":
		userData := fsmManager.GetData(chatID)
		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("◀️ Возврат назад\n\n"+
				"✅ Название: %s\n\n"+
				"💰 Шаг 3/4: Введи бюджет (в рублях):",
				userData.ProjectName))
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboards.BackButton("back_to_name")
		bot.Send(msg)

	// Выбор мастера (динамически из БД по ID) — создание, смена или выбор
	case strings.HasPrefix(data, "master_pick_") || data == "master_none":
		var masterName string

		if data == "master_none" {
			masterName = "Не назначен"
		} else {
			masterIDStr := strings.TrimPrefix(data, "master_pick_")
			masterID, err := strconv.ParseInt(masterIDStr, 10, 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID мастера"))
				return
			}
			master, err := database.GetMasterByID(masterID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Мастер не найден"))
				return
			}
			masterName = master.Name
		}

		userData := fsmManager.GetData(chatID)
		userData.ProjectMaster = masterName
		state := fsmManager.GetState(chatID)

		switch {
		// Контекст 1: создание нового проекта (шаг 4/4)
		case state == fsm.StateCreatingProjectMaster:
			fsmManager.SetData(chatID, userData)

			budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)
			projectID, err := database.CreateProject(
				chatID,
				userData.ProjectType,
				userData.ProjectName,
				budget,
				masterName,
			)
			if err != nil {
				log.Printf("❌ Ошибка сохранения проекта: %v", err)
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка сохранения проекта в БД"))
				return
			}
			log.Printf("✅ Проект ID=%d сохранён в БД", projectID)

			// Сбрасываем только состояние (данные оставляем для "✏️ Редактировать"),
			// чтобы повторное открытие клавиатуры мастеров не создало дубль проекта.
			fsmManager.SetState(chatID, fsm.StateIdle)

			typeEmoji := getProjectTypeEmoji(userData.ProjectType)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
				"🎉 Проект создан!\n\n"+
					"%s Тип: %s\n"+
					"📋 Название: %s\n"+
					"💰 Бюджет: %s ₽\n"+
					"👷 Ответственный: %s\n"+
					"📅 Дата создания: %s\n"+
					"📍 Статус: В работе\n\n"+
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

		// Контекст 2: смена мастера у существующего проекта
		case userData.EditingProjectID != 0:
			project, err := database.GetProjectByID(userData.EditingProjectID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
				fsmManager.ResetState(chatID)
				return
			}

			if err := database.UpdateProject(project.ID, project.Name, project.Budget, masterName); err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка обновления мастера"))
				return
			}

			userData.EditingProjectID = 0
			fsmManager.SetData(chatID, userData)

			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Ответственный обновлён: 👷 %s", masterName)))

		// Контекст 3: выбор во временном (in-memory) редактировании
		default:
			fsmManager.SetData(chatID, userData)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("👷 Ответственный выбран: %s", masterName)))
		}

	case data == "edit_last_project":
		userData := fsmManager.GetData(chatID)
		if userData.ProjectName == "" {
			msg := tgbotapi.NewMessage(chatID, "❌ Нет данных проекта для редактирования.")
			bot.Send(msg)
			return
		}

		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✏️ Редактирование проекта\n\n"+
				"Текущие данные:\n"+
				"%s Тип: %s\n"+
				"📋 Название: %s\n"+
				"💰 Бюджет: %s ₽\n"+
				"👷 Ответственный: %s\n\n"+
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

	// Редактирование названия
	case strings.HasPrefix(data, "edit_name_"):
		projectIDStr := strings.TrimPrefix(data, "edit_name_")
		projectID, _ := strconv.ParseInt(projectIDStr, 10, 64)

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateEditingProjectName)

		msg := tgbotapi.NewMessage(chatID,
			"📋 *Изменение названия*\n\n"+
				"Введи новое название проекта:")
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	// Редактирование бюджета
	case strings.HasPrefix(data, "edit_budget_"):
		projectIDStr := strings.TrimPrefix(data, "edit_budget_")
		projectID, _ := strconv.ParseInt(projectIDStr, 10, 64)

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateEditingProjectBudget)

		msg := tgbotapi.NewMessage(chatID,
			"💰 *Изменение бюджета*\n\n"+
				"Введи новый бюджет (в рублях):")
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	// Редактирование ФИО мастера (раздел "👷 Мастера")
	case strings.HasPrefix(data, "edit_master_name_"):
		masterIDStr := strings.TrimPrefix(data, "edit_master_name_")
		masterID, err := strconv.ParseInt(masterIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID мастера"))
			return
		}

		master, err := database.GetMasterByID(masterID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Мастер не найден"))
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.EditingMasterID = masterID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateEditingMasterName)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✏️ *Изменение ФИО мастера*\n\n"+
				"Текущее ФИО: *%s*\n\n"+
				"Введи новое ФИО:", master.Name))
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	// Редактирование мастера у проекта
	case strings.HasPrefix(data, "edit_master_"):
		projectIDStr := strings.TrimPrefix(data, "edit_master_")
		projectID, _ := strconv.ParseInt(projectIDStr, 10, 64)

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		fsmManager.SetData(chatID, userData)

		msg := tgbotapi.NewMessage(chatID,
			"👷 *Изменение мастера*\n\n"+
				"Выбери нового мастера:")
		msg.ReplyMarkup = keyboards.MastersKeyboard()
		bot.Send(msg)

	case data == "edit_type":
		msg := tgbotapi.NewMessage(chatID, "🔧 Выбери новый тип проекта:")
		msg.ReplyMarkup = keyboards.ProjectTypeKeyboard()
		bot.Send(msg)

	case data == "edit_name":
		fsmManager.SetState(chatID, fsm.StateCreatingProject)
		msg := tgbotapi.NewMessage(chatID, "📋 Введи новое название проекта:")
		bot.Send(msg)

	case data == "edit_budget":
		fsmManager.SetState(chatID, fsm.StateCreatingProjectBudget)
		msg := tgbotapi.NewMessage(chatID, "💰 Введи новый бюджет:")
		bot.Send(msg)

	case data == "edit_master":
		msg := tgbotapi.NewMessage(chatID, "👷 Выбери нового ответственного:")
		msg.ReplyMarkup = keyboards.MastersKeyboard()
		bot.Send(msg)

	case data == "save_project":
		userData := fsmManager.GetData(chatID)
		log.Printf("📝 SAVE: Тип=%s, Название=%s, Бюджет=%s, Мастер=%s",
			userData.ProjectType,
			userData.ProjectName,
			userData.ProjectBudget,
			userData.ProjectMaster)

		typeEmoji := getProjectTypeEmoji(userData.ProjectType)
		budget, _ := strconv.ParseFloat(userData.ProjectBudget, 64)

		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"✅ Проект сохранён!\n\n"+
				"%s Тип: %s\n"+
				"📋 Название: %s\n"+
				"💰 Бюджет: %s ₽\n"+
				"👷 Ответственный: %s\n"+
				"📅 Обновлён: %s",
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

	case data == "back_menu":
		msg := tgbotapi.NewMessage(chatID, "◀️ Возврат в главное меню")
		bot.Send(msg)
		fsmManager.ResetState(chatID)

	// Редактирование проекта
	case strings.HasPrefix(data, "edit_project_"):
		projectIDStr := strings.TrimPrefix(data, "edit_project_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID проекта")
			bot.Send(msg)
			return
		}

		project, err := database.GetProjectByID(projectID)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Проект не найден")
			bot.Send(msg)
			return
		}

		masterName := database.GetMasterNameByID(project.MasterID)

		text := fmt.Sprintf(
			"✏️ *Редактирование проекта*\n\n"+
				"*%s*\n\n"+
				"Текущие данные:\n"+
				"🔧 Тип: %s\n"+
				"💰 Бюджет: %.2f ₽\n"+
				"👷 Мастер: %s\n\n"+
				"Что хочешь изменить?",
			project.Name,
			project.Type,
			project.Budget,
			masterName,
		)

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📋 Название", fmt.Sprintf("edit_name_%d", projectID)),
				tgbotapi.NewInlineKeyboardButtonData("💰 Бюджет", fmt.Sprintf("edit_budget_%d", projectID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("👷 Мастера", fmt.Sprintf("edit_master_%d", projectID)),
				tgbotapi.NewInlineKeyboardButtonData("💰 Расходы", fmt.Sprintf("expenses_%d", projectID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏗 Этапы", fmt.Sprintf("stages_%d", projectID)),
				tgbotapi.NewInlineKeyboardButtonData("📇 Карточка объекта", fmt.Sprintf("cardmenu_%d", projectID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "back_to_projects"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)

	// Карточка объекта/заказчика
	case strings.HasPrefix(data, "cardmenu_"):
		projectIDStr := strings.TrimPrefix(data, "cardmenu_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if _, err := database.GetProjectByID(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}
		sendCardMenu(bot, chatID, projectID)

	// Старт редактирования поля карточки: cardset_<field>_<projectID>
	case strings.HasPrefix(data, "cardset_"):
		rest := strings.TrimPrefix(data, "cardset_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный формат"))
			return
		}
		field := parts[0]
		projectID, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if _, err := database.GetProjectByID(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		userData.CardField = field
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateEditingProjectCard)

		bot.Send(tgbotapi.NewMessage(chatID, cardFieldPrompt(field)))

	// Меню этапов проекта (вехи)
	case strings.HasPrefix(data, "stages_"):
		projectIDStr := strings.TrimPrefix(data, "stages_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if _, err := database.GetProjectByID(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}

		sendStagesMenu(bot, chatID, projectID)

	// Добавление этапа
	case strings.HasPrefix(data, "add_stage_"):
		projectIDStr := strings.TrimPrefix(data, "add_stage_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if _, err := database.GetProjectByID(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateAddingStageName)

		bot.Send(tgbotapi.NewMessage(chatID, "🏗 Введи название этапа:"))

	// Типовые этапы
	case strings.HasPrefix(data, "seed_stages_"):
		projectIDStr := strings.TrimPrefix(data, "seed_stages_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if err := database.SeedStandardStages(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось добавить этапы"))
			return
		}
		sendStagesMenu(bot, chatID, projectID)

	// Переключение статуса этапа (цикл)
	case strings.HasPrefix(data, "stage_next_"):
		stageIDStr := strings.TrimPrefix(data, "stage_next_")
		stageID, err := strconv.ParseInt(stageIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID этапа"))
			return
		}
		stage, err := database.GetStageByID(stageID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Этап не найден"))
			return
		}
		if err := database.UpdateStageStatus(stageID, database.NextStageStatus(stage.Status)); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось изменить этап"))
			return
		}
		sendStagesMenu(bot, chatID, stage.ProjectID)

	// Меню расходов проекта (план/факт)
	case strings.HasPrefix(data, "expenses_"):
		projectIDStr := strings.TrimPrefix(data, "expenses_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}

		status, err := database.GetProjectBudgetStatus(projectID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}

		warn := ""
		if status.Over {
			warn = "\n🔴 *Перерасход бюджета!*"
		}

		text := fmt.Sprintf(
			"💰 *Бюджет проекта (план/факт)*\n\n"+
				"📊 План: %.2f ₽\n"+
				"💸 Освоено: %.2f ₽ (%d%%)\n"+
				"💵 Остаток: %.2f ₽%s\n",
			status.Budget, status.Spent, status.Percent, status.Remaining, warn)

		expenses, _ := database.GetProjectExpenses(projectID)
		if len(expenses) > 0 {
			text += "\n*Последние расходы:*\n"
			limit := 5
			for i, e := range expenses {
				if i >= limit {
					break
				}
				descr := e.Description
				if descr == "" {
					descr = "без описания"
				}
				text += fmt.Sprintf("• %.2f ₽ — %s\n", e.Amount, descr)
			}
		}

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить расход", fmt.Sprintf("add_expense_%d", projectID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ К проекту", fmt.Sprintf("edit_project_%d", projectID)),
			),
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)

	// Старт добавления расхода
	case strings.HasPrefix(data, "add_expense_"):
		projectIDStr := strings.TrimPrefix(data, "add_expense_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID проекта"))
			return
		}
		if _, err := database.GetProjectByID(projectID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.EditingProjectID = projectID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateAddingExpenseAmount)

		bot.Send(tgbotapi.NewMessage(chatID, "💸 Введи сумму расхода (₽):"))

	// Удаление проекта (с подтверждением)
	case strings.HasPrefix(data, "delete_project_"):
		projectIDStr := strings.TrimPrefix(data, "delete_project_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID проекта")
			bot.Send(msg)
			return
		}

		project, err := database.GetProjectByID(projectID)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Проект не найден")
			bot.Send(msg)
			return
		}

		text := fmt.Sprintf(
			"⚠️ *Подтверждение удаления*\n\n"+
				"Точно удалить проект?\n\n"+
				"*%s*\n"+
				"💰 Бюджет: %.2f ₽\n\n"+
				"⚠️ Это действие нельзя отменить!",
			project.Name,
			project.Budget,
		)

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить", fmt.Sprintf("confirm_delete_%d", projectID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "back_to_projects"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)

	// Подтверждённое удаление
	case strings.HasPrefix(data, "confirm_delete_"):
		projectIDStr := strings.TrimPrefix(data, "confirm_delete_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID проекта")
			bot.Send(msg)
			return
		}

		err = database.DeleteProject(projectID)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка удаления проекта")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID,
			"✅ *Проект удалён*\n\n"+
				"Проект успешно удалён из базы данных.")
		msg.ParseMode = "Markdown"
		bot.Send(msg)

	// Возврат к списку проектов
	case data == "back_to_projects":
		msg := tgbotapi.NewMessage(chatID, "Используй /myprojects для просмотра проектов")
		bot.Send(msg)

		// Просмотр задач проекта
	case strings.HasPrefix(data, "view_tasks_"):
		projectIDStr := strings.TrimPrefix(data, "view_tasks_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID проекта")
			bot.Send(msg)
			return
		}

		project, err := database.GetProjectByID(projectID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Проект не найден")
			bot.Send(msg)
			return
		}

		tasks, err := database.GetProjectTasks(projectID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки задач")
			bot.Send(msg)
			return
		}

		if len(tasks) == 0 {
			text := fmt.Sprintf(
				"📝 *Задачи проекта*\n\n"+
					"*%s*\n\n"+
					"📋 Пока нет задач\n\n"+
					"➕ Добавь первую задачу!",
				project.Name,
			)

			buttons := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("➕ Добавить задачу", fmt.Sprintf("add_task_%d", projectID)),
					tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "back_to_projects"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = buttons
			bot.Send(msg)
			return
		}

		// Отправляем каждую задачу отдельным сообщением с кнопками
		for i, task := range tasks {
			statusEmoji := database.GetTaskStatusEmoji(task.Status)
			statusName := database.GetTaskStatusName(task.Status)

			deadlineText := "Не указан"
			if task.Deadline != nil {
				deadlineText = task.Deadline.Format("02.01.2006")
				if database.IsTaskOverdue(task) {
					deadlineText += " 🔴 ПРОСРОЧЕНО"
				}
			}

			text := fmt.Sprintf(
				"📝 *Задача %d из %d*\n\n"+
					"*%s* %s\n\n"+
					"📄 %s\n"+
					"📅 Дедлайн: %s\n"+
					"👷 Исполнитель: %s\n"+
					"⏱ Статус: %s",
				i+1, len(tasks),
				task.Name, statusEmoji,
				task.Description,
				deadlineText,
				database.GetMasterNameByID(task.AssigneeID),
				statusName,
			)

			// Кнопки в зависимости от статуса
			var statusButtons []tgbotapi.InlineKeyboardButton

			switch task.Status {
			case "pending":
				statusButtons = []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("▶️ Начать", fmt.Sprintf("task_status_%d_in_progress", task.ID)),
				}
			case "in_progress":
				statusButtons = []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("🔎 Сдать на проверку", fmt.Sprintf("task_status_%d_review", task.ID)),
				}
			case "review":
				statusButtons = []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("✅ Принять", fmt.Sprintf("task_accept_%d", task.ID)),
					tgbotapi.NewInlineKeyboardButtonData("🔁 На доработку", fmt.Sprintf("task_reject_%d", task.ID)),
				}
			case "completed":
				statusButtons = []tgbotapi.InlineKeyboardButton{
					tgbotapi.NewInlineKeyboardButtonData("🔄 Вернуть", fmt.Sprintf("task_status_%d_in_progress", task.ID)),
				}
			}

			buttons := tgbotapi.NewInlineKeyboardMarkup(
				statusButtons,
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать", fmt.Sprintf("edit_task_%d", task.ID)),
					tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить", fmt.Sprintf("delete_task_%d", task.ID)),
				),
			)

			msg := tgbotapi.NewMessage(chatID, text)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = buttons
			bot.Send(msg)
		}

		// Прогресс проекта в процентах
		completedCount := 0
		for _, t := range tasks {
			if t.Status == "completed" {
				completedCount++
			}
		}
		progress := database.CalcProgress(len(tasks), completedCount)

		// Кнопки после списка задач
		finalButtons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить задачу", fmt.Sprintf("add_task_%d", projectID)),
				tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "back_to_projects"),
			),
		)

		finalMsg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"📊 Прогресс проекта: %s %d%% (%d/%d)",
			progressBar(progress), progress, completedCount, len(tasks)))
		finalMsg.ReplyMarkup = finalButtons
		bot.Send(finalMsg)

	// Создание задачи
	case strings.HasPrefix(data, "add_task_"):
		projectIDStr := strings.TrimPrefix(data, "add_task_")
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)

		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка: неверный ID проекта")
			bot.Send(msg)
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.TaskProjectID = projectID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateCreatingTaskName)

		msg := tgbotapi.NewMessage(chatID,
			"➕ *Новая задача*\n\n"+
				"📝 Введи название задачи:")
		msg.ParseMode = "Markdown"
		bot.Send(msg)

		// Изменение статуса задачи
	case strings.HasPrefix(data, "task_status_"):
		// Формат: task_status_{ID}_{STATUS}
		// Статус может содержать underscore (in_progress)
		rest := strings.TrimPrefix(data, "task_status_")

		// Ищем первый underscore - это разделитель ID и статуса
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) < 2 {
			msg := tgbotapi.NewMessage(chatID, "❌ Неверный формат callback")
			bot.Send(msg)
			return
		}

		taskID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи")
			bot.Send(msg)
			return
		}

		newStatus := parts[1] // "in_progress", "pending", "completed"

		err = database.UpdateTaskStatus(taskID, newStatus)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("❌ Ошибка: %v", err))
			bot.Send(msg)
			return
		}

		statusEmoji := database.GetTaskStatusEmoji(newStatus)
		statusName := database.GetTaskStatusName(newStatus)

		msg := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Статус изменён: %s %s", statusEmoji, statusName))
		bot.Send(msg)

	// Приёмка работ: принять задачу (только из статуса «на проверке»)
	case strings.HasPrefix(data, "task_accept_"):
		taskID, err := strconv.ParseInt(strings.TrimPrefix(data, "task_accept_"), 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}
		task, err := database.GetTaskByID(taskID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}
		if task.Status != "review" {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Задача не на проверке — принимать нечего"))
			return
		}
		if err := database.UpdateTaskStatus(taskID, "completed"); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка приёмки"))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, "✅ Работа принята"))

	// Приёмка работ: вернуть на доработку
	case strings.HasPrefix(data, "task_reject_"):
		taskID, err := strconv.ParseInt(strings.TrimPrefix(data, "task_reject_"), 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}
		task, err := database.GetTaskByID(taskID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}
		if task.Status != "review" {
			bot.Send(tgbotapi.NewMessage(chatID, "⚠️ Задача не на проверке"))
			return
		}
		if err := database.UpdateTaskStatus(taskID, "in_progress"); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка возврата"))
			return
		}
		bot.Send(tgbotapi.NewMessage(chatID, "🔁 Задача возвращена на доработку"))

	// Удаление задачи
	case strings.HasPrefix(data, "delete_task_"):
		taskIDStr := strings.TrimPrefix(data, "delete_task_")
		taskID, _ := strconv.ParseInt(taskIDStr, 10, 64)

		err := database.DeleteTask(taskID)
		if err != nil {
			msg := tgbotapi.NewMessage(chatID, "❌ Ошибка удаления задачи")
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(chatID, "✅ Задача удалена")
		bot.Send(msg)

	// Редактирование названия задачи (специфичные префиксы — перед edit_task_)
	case strings.HasPrefix(data, "edit_task_name_"):
		startTaskFieldEdit(bot, chatID, fsmManager, data, "edit_task_name_", fsm.StateEditingTaskName,
			"📝 Введи новое название задачи:")

	case strings.HasPrefix(data, "edit_task_desc_"):
		startTaskFieldEdit(bot, chatID, fsmManager, data, "edit_task_desc_", fsm.StateEditingTaskDescription,
			"📄 Введи новое описание задачи (или /clear — очистить):")

	case strings.HasPrefix(data, "edit_task_deadline_"):
		startTaskFieldEdit(bot, chatID, fsmManager, data, "edit_task_deadline_", fsm.StateEditingTaskDeadline,
			"📅 Введи новый срок в формате ДД.ММ.ГГГГ (или /clear — убрать срок):")

	// Выбор исполнителя задачи — показываем клавиатуру мастеров
	case strings.HasPrefix(data, "edit_task_assignee_"):
		taskIDStr := strings.TrimPrefix(data, "edit_task_assignee_")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}
		if _, err := database.GetTaskByID(taskID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, "👷 Выбери исполнителя задачи:")
		msg.ReplyMarkup = keyboards.TaskAssigneeKeyboard(taskID)
		bot.Send(msg)

	// Назначение исполнителя: task_assignee_<taskID>_<masterID> (0 — снять)
	case strings.HasPrefix(data, "task_assignee_"):
		rest := strings.TrimPrefix(data, "task_assignee_")
		parts := strings.SplitN(rest, "_", 2)
		if len(parts) < 2 {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный формат"))
			return
		}
		taskID, err1 := strconv.ParseInt(parts[0], 10, 64)
		masterID, err2 := strconv.ParseInt(parts[1], 10, 64)
		if err1 != nil || err2 != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID"))
			return
		}

		var assignee *int64
		assigneeName := "Не назначен"
		if masterID != 0 {
			master, err := database.GetMasterByID(masterID)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "❌ Мастер не найден"))
				return
			}
			assignee = &masterID
			assigneeName = master.Name
		}

		if err := database.UpdateTaskAssignee(taskID, assignee); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка назначения исполнителя"))
			return
		}

		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("✅ Исполнитель задачи: 👷 %s", assigneeName)))

	// Меню редактирования задачи
	case strings.HasPrefix(data, "edit_task_"):
		taskIDStr := strings.TrimPrefix(data, "edit_task_")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}

		task, err := database.GetTaskByID(taskID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}

		deadlineText := "не указан"
		if task.Deadline != nil {
			deadlineText = task.Deadline.Format("02.01.2006")
		}
		assigneeName := database.GetMasterNameByID(task.AssigneeID)

		text := fmt.Sprintf(
			"✏️ *Редактирование задачи*\n\n"+
				"📝 %s\n"+
				"📄 %s\n"+
				"📅 Срок: %s\n"+
				"👷 Исполнитель: %s\n\n"+
				"Что изменить?",
			task.Name, task.Description, deadlineText, assigneeName)

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📝 Название", fmt.Sprintf("edit_task_name_%d", taskID)),
				tgbotapi.NewInlineKeyboardButtonData("📄 Описание", fmt.Sprintf("edit_task_desc_%d", taskID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📅 Срок", fmt.Sprintf("edit_task_deadline_%d", taskID)),
				tgbotapi.NewInlineKeyboardButtonData("👷 Исполнитель", fmt.Sprintf("edit_task_assignee_%d", taskID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📷 Фото работ", fmt.Sprintf("task_photos_%d", taskID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ К задачам", fmt.Sprintf("view_tasks_%d", task.ProjectID)),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)

	// Меню фотофиксации задачи
	case strings.HasPrefix(data, "task_photos_"):
		taskIDStr := strings.TrimPrefix(data, "task_photos_")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}
		if _, err := database.GetTaskByID(taskID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}

		count, _ := database.CountTaskPhotos(taskID)
		text := fmt.Sprintf("📷 *Фотофиксация работ*\n\nПрикреплено фото: %d", count)

		buttons := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить фото", fmt.Sprintf("add_photo_%d", taskID)),
				tgbotapi.NewInlineKeyboardButtonData("🖼 Показать", fmt.Sprintf("show_photos_%d", taskID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ К задаче", fmt.Sprintf("edit_task_%d", taskID)),
			),
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buttons
		bot.Send(msg)

	// Старт добавления фото — ждём изображение
	case strings.HasPrefix(data, "add_photo_"):
		taskIDStr := strings.TrimPrefix(data, "add_photo_")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}
		if _, err := database.GetTaskByID(taskID); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
			return
		}

		userData := fsmManager.GetData(chatID)
		userData.EditingTaskID = taskID
		fsmManager.SetData(chatID, userData)
		fsmManager.SetState(chatID, fsm.StateAddingTaskPhoto)

		bot.Send(tgbotapi.NewMessage(chatID,
			"📷 Отправь фото работ (можно с подписью). /cancel — отмена."))

	// Показать фотофиксацию задачи
	case strings.HasPrefix(data, "show_photos_"):
		taskIDStr := strings.TrimPrefix(data, "show_photos_")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
			return
		}

		photos, err := database.GetTaskPhotos(taskID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка загрузки фото"))
			return
		}
		if len(photos) == 0 {
			bot.Send(tgbotapi.NewMessage(chatID, "📷 Фото пока нет."))
			return
		}

		for _, p := range photos {
			photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(p.FileID))
			if p.Caption != "" {
				photoMsg.Caption = p.Caption
			}
			bot.Send(photoMsg)
		}

	// Переключение уведомлений
	case data == "toggle_notifications":
		settings, err := database.GetUserSettings(chatID)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка чтения настроек"))
			return
		}
		newValue := !settings.Notifications
		if err := database.SetNotifications(chatID, newValue); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось сохранить настройку"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, notificationsText(newValue))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboards.NotificationsKeyboard(newValue)
		bot.Send(msg)

	// Смена роли пользователя
	case strings.HasPrefix(data, "set_role_"):
		role := strings.TrimPrefix(data, "set_role_")
		if !database.IsValidRole(role) {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Неизвестная роль"))
			return
		}
		if err := database.SetUserRole(chatID, role); err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось сменить роль"))
			return
		}

		msg := tgbotapi.NewMessage(chatID, roleScreenText(role))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = keyboards.RolesKeyboard(role)
		bot.Send(msg)

	default:
		log.Printf("⚠️ Неизвестный callback_data: %s", data)
		msg := tgbotapi.NewMessage(chatID, "❌ Неизвестная команда")
		bot.Send(msg)
	}
}

// startTaskFieldEdit разбирает ID задачи из callback, переводит пользователя
// в нужное состояние редактирования и отправляет приглашение к вводу.
func startTaskFieldEdit(bot *tgbotapi.BotAPI, chatID int64, fsmManager *fsm.Manager,
	data, prefix string, state fsm.State, prompt string) {

	taskIDStr := strings.TrimPrefix(data, prefix)
	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Неверный ID задачи"))
		return
	}

	if _, err := database.GetTaskByID(taskID); err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Задача не найдена"))
		return
	}

	userData := fsmManager.GetData(chatID)
	userData.EditingTaskID = taskID
	fsmManager.SetData(chatID, userData)
	fsmManager.SetState(chatID, state)

	bot.Send(tgbotapi.NewMessage(chatID, prompt))
}

// notificationsText — формирует текст экрана уведомлений по текущему состоянию.
func notificationsText(on bool) string {
	status := "❌ Выключены"
	if on {
		status = "✅ Включены"
	}
	return fmt.Sprintf("🔔 *Уведомления*\n\nТекущее состояние: %s\n\n"+
		"Уведомления о сроках и просрочках задач.", status)
}

// ИСПРАВЛЕНО: убрал "data ==" из case
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

// cardFieldPrompt — приглашение к вводу для поля карточки.
func cardFieldPrompt(field string) string {
	switch field {
	case "address":
		return "📍 Введи адрес объекта (или /clear):"
	case "customer":
		return "🧑‍💼 Введи заказчика (или /clear):"
	case "contract":
		return "📄 Введи № договора (или /clear):"
	case "start":
		return "📅 Введи дату старта ДД.ММ.ГГГГ (или /clear):"
	case "end":
		return "🏁 Введи плановую дату сдачи ДД.ММ.ГГГГ (или /clear):"
	default:
		return "Введи значение:"
	}
}

// cardValueOrDash — значение поля или прочерк.
func cardValueOrDash(v string) string {
	if v == "" {
		return "—"
	}
	return v
}

// cardDateOrDash — дата в формате ДД.ММ.ГГГГ или прочерк.
func cardDateOrDash(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("02.01.2006")
}

// sendCardMenu отправляет карточку объекта/заказчика с кнопками правки полей.
func sendCardMenu(bot *tgbotapi.BotAPI, chatID, projectID int64) {
	p, err := database.GetProjectByID(projectID)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Проект не найден"))
		return
	}

	text := fmt.Sprintf(
		"📇 *Карточка объекта*\n\n"+
			"*%s*\n\n"+
			"📍 Адрес: %s\n"+
			"🧑‍💼 Заказчик: %s\n"+
			"📄 Договор: %s\n"+
			"📅 Старт: %s\n"+
			"🏁 Плановая сдача: %s",
		p.Name,
		cardValueOrDash(p.Address),
		cardValueOrDash(p.Customer),
		cardValueOrDash(p.ContractNumber),
		cardDateOrDash(p.StartDate),
		cardDateOrDash(p.PlannedEnd),
	)

	buttons := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📍 Адрес", fmt.Sprintf("cardset_address_%d", projectID)),
			tgbotapi.NewInlineKeyboardButtonData("🧑‍💼 Заказчик", fmt.Sprintf("cardset_customer_%d", projectID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Договор", fmt.Sprintf("cardset_contract_%d", projectID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Старт", fmt.Sprintf("cardset_start_%d", projectID)),
			tgbotapi.NewInlineKeyboardButtonData("🏁 Сдача", fmt.Sprintf("cardset_end_%d", projectID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ К проекту", fmt.Sprintf("edit_project_%d", projectID)),
		),
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = buttons
	bot.Send(msg)
}

// stageStatusEmoji — значок статуса этапа.
func stageStatusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "in_progress":
		return "⚙️"
	default:
		return "🕐"
	}
}

// sendStagesMenu отправляет меню этапов проекта с прогрессом по вехам.
func sendStagesMenu(bot *tgbotapi.BotAPI, chatID, projectID int64) {
	stages, _ := database.GetProjectStages(projectID)
	percent, total, completed, _ := database.GetProjectStageProgress(projectID)

	text := "🏗 *Этапы проекта*\n\n"
	if total == 0 {
		text += "Этапов пока нет.\nДобавь свой или подставь типовые."
	} else {
		for _, s := range stages {
			text += fmt.Sprintf("%s %s\n", stageStatusEmoji(s.Status), s.Name)
		}
		text += fmt.Sprintf("\n📊 Прогресс по этапам: %s %d%% (%d/%d)",
			progressBar(percent), percent, completed, total)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range stages {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s", stageStatusEmoji(s.Status), s.Name),
				fmt.Sprintf("stage_next_%d", s.ID)),
		))
	}

	addRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("➕ Этап", fmt.Sprintf("add_stage_%d", projectID)),
	}
	if total == 0 {
		addRow = append(addRow,
			tgbotapi.NewInlineKeyboardButtonData("🏗 Типовые этапы", fmt.Sprintf("seed_stages_%d", projectID)))
	}
	rows = append(rows, addRow)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("◀️ К проекту", fmt.Sprintf("edit_project_%d", projectID)),
	))

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	bot.Send(msg)
}

// progressBar — визуальная полоса прогресса из 10 сегментов (▰ заполнено, ▱ пусто)
func progressBar(percent int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := percent / 10
	return strings.Repeat("▰", filled) + strings.Repeat("▱", 10-filled)
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
