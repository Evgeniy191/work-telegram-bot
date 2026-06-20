package handlers

import (
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestMenuHandlers(t *testing.T) {
	handlers := []struct {
		name string
		fn   func(*tgbotapi.BotAPI, tgbotapi.Update)
	}{
		{"HandleProjects", HandleProjects},
		{"HandleTasks", HandleTasks},
		{"HandleReports", HandleReports},
		{"HandleSettings", HandleSettings},
		{"HandleSettingsMenu", HandleSettingsMenu},
		{"HandleQuickActions", HandleQuickActions},
		{"HandleNewProject", HandleNewProject},
		{"HandleNewTask", HandleNewTask},
		{"HandleMyTasks", HandleMyTasks},
		{"HandleLanguage", HandleLanguage},
		{"HandleNotifications", HandleNotifications},
		{"HandleTheme", HandleTheme},
		{"HandleSecurity", HandleSecurity},
		{"HandleReportFormat", HandleReportFormat},
		{"HandleStart", HandleStart},
		{"HandleHelp", HandleHelp},
		{"HandleAbout", HandleAbout},
	}

	for _, h := range handlers {
		t.Run(h.name, func(t *testing.T) {
			bot, cap := newTestBot(t)
			h.fn(bot, msgUpdate(1, "x"))
			if cap.sends() == 0 {
				t.Errorf("%s не отправил ни одного сообщения", h.name)
			}
		})
	}
}

func TestHandleMasters(t *testing.T) {
	bot, cap := newTestBot(t)
	HandleMasters(bot, msgUpdate(1, "👷 Мастера"))
	if !cap.contains("Управление мастерами") {
		t.Errorf("ожидался список мастеров, получили: %s", cap.joined())
	}
}

func TestHandleMyProjectsEmpty(t *testing.T) {
	bot, cap := newTestBot(t)
	// Пользователь без проектов
	HandleMyProjects(bot, msgUpdate(7777, "📋 Проекты"))
	if !cap.contains("нет проектов") {
		t.Errorf("ожидалось сообщение об отсутствии проектов, получили: %s", cap.joined())
	}
}

func TestHandleMyProjectsWithProgress(t *testing.T) {
	bot, cap := newTestBot(t)
	chatID := int64(8888)

	// Готовим проект с задачами
	pid, err := dbCreateProjectWithTasks(chatID, 4, 1)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	_ = pid

	HandleMyProjects(bot, msgUpdate(chatID, "📋 Проекты"))
	if !cap.contains("Прогресс") {
		t.Errorf("ожидался прогресс в карточке проекта, получили: %s", cap.joined())
	}
	if !cap.contains("25%") {
		t.Errorf("ожидался прогресс 25%%, получили: %s", cap.joined())
	}
}

func TestProcessUpdateCallback(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	ProcessUpdate(bot, cbUpdate(1, "back_menu"), mgr)
	if cap.sends() == 0 {
		t.Error("callback не обработан")
	}
}

func TestProcessUpdateNilMessage(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	// Update без Message и без CallbackQuery — должно тихо игнорироваться
	ProcessUpdate(bot, tgbotapi.Update{}, mgr)
	if cap.sends() != 0 {
		t.Error("пустой update не должен слать сообщения")
	}
}

func TestProcessUpdateCancel(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(11)

	// Нечего отменять
	ProcessUpdate(bot, msgUpdate(chatID, "/cancel"), mgr)
	if !cap.contains("Нечего отменять") {
		t.Errorf("ожидалось 'Нечего отменять', получили: %s", cap.joined())
	}

	// Есть что отменять
	mgr.SetState(chatID, fsm.StateCreatingProject)
	ProcessUpdate(bot, msgUpdate(chatID, "/cancel"), mgr)
	if !cap.contains("отменено") {
		t.Errorf("ожидалось 'отменено', получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("состояние не сброшено после /cancel")
	}
}

func TestProcessUpdateFSMIntercept(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(12)
	mgr.SetState(chatID, fsm.StateCreatingProject)

	ProcessUpdate(bot, msgUpdate(chatID, "Мой проект"), mgr)
	// FSM должен перехватить и попросить бюджет
	if !cap.contains("бюджет") && !cap.contains("Бюджет") {
		t.Errorf("FSM не перехватил сообщение, получили: %s", cap.joined())
	}
}

func TestRouteCommandAllButtons(t *testing.T) {
	mgr := fsm.NewManager()
	buttons := []string{
		"/start", "/help", "❓ Помощь", "/about",
		"📋 Проекты", "/myprojects", "📋 Мои проекты",
		"➕ Новый проект", "➕ Новая задача",
		"📝 Задачи", "👷 Мастера", "📊 Отчёты", "⚙️ Настройки",
		"🏠 Главное меню", "⚡ Быстрые действия", "📝 Мои задачи",
		"🌍 Язык", "🔔 Уведомления", "🎨 Тема", "🔐 Безопасность",
		"📊 Формат отчётов",
	}
	for _, b := range buttons {
		t.Run(b, func(t *testing.T) {
			bot, cap := newTestBot(t)
			RouteCommand(bot, msgUpdate(100+int64(len(b)), b), mgr)
			if cap.contains("Я не понял") {
				t.Errorf("кнопка %q не распознана", b)
			}
			if cap.sends() == 0 {
				t.Errorf("кнопка %q не вызвала ответа", b)
			}
		})
	}
}

func TestRouteCommandUnknown(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	RouteCommand(bot, msgUpdate(1, "абракадабра"), mgr)
	if !cap.contains("Я не понял") {
		t.Errorf("ожидалось 'Я не понял', получили: %s", cap.joined())
	}
}
