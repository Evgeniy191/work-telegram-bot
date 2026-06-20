package handlers

import (
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestFSMCreatingProjectName(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(2001)
	mgr.SetState(chatID, fsm.StateCreatingProject)

	if !HandleFSMMessage(bot, msgUpdate(chatID, "Проект X"), mgr) {
		t.Fatal("ожидалось, что FSM обработает сообщение")
	}
	if mgr.GetState(chatID) != fsm.StateCreatingProjectBudget {
		t.Errorf("ожидался переход к бюджету, состояние=%s", mgr.GetState(chatID))
	}
	if mgr.GetData(chatID).ProjectName != "Проект X" {
		t.Error("имя проекта не сохранено")
	}
	_ = cap
}

func TestFSMBudgetValidation(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2002)

	// Не число
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateCreatingProjectBudget)
	HandleFSMMessage(bot, msgUpdate(chatID, "абв"), mgr)
	if !cap.contains("числом") {
		t.Errorf("ожидалась ошибка валидации, получили: %s", cap.joined())
	}

	// Отрицательный
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "-5"), mgr)
	if !cap.contains("отрицательным") {
		t.Errorf("ожидалась ошибка отрицательного, получили: %s", cap.joined())
	}

	// Корректный
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "12500.50"), mgr)
	if mgr.GetState(chatID) != fsm.StateCreatingProjectMaster {
		t.Errorf("ожидался переход к выбору мастера, состояние=%s", mgr.GetState(chatID))
	}
	if !cap.contains("ответственного") {
		t.Errorf("ожидался запрос мастера, получили: %s", cap.joined())
	}
}

func TestFSMCreatingTask(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(2003)
	mgr.SetState(chatID, fsm.StateCreatingTask)

	HandleFSMMessage(bot, msgUpdate(chatID, "Простая задача"), mgr)
	if !cap.contains("Задача создана") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("состояние должно сброситься")
	}
}

func TestFSMIdleAndUnknown(t *testing.T) {
	bot, _ := newTestBot(t)
	mgr := fsm.NewManager()

	if HandleFSMMessage(bot, msgUpdate(2004, "привет"), mgr) {
		t.Error("в StateIdle FSM не должен перехватывать сообщение")
	}

	mgr.SetState(2005, fsm.StateEditingTaskDeadline) // состояние без обработчика
	if HandleFSMMessage(bot, msgUpdate(2005, "x"), mgr) {
		t.Error("для необработанного состояния ожидался false")
	}
}

func TestFSMEditingProjectName(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2006)

	pid, err := database.CreateProject(chatID, "Монтаж", "Старое", 100, "Не назначен")
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Пустое имя
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingProjectName)
	mgr.GetData(chatID).EditingProjectID = pid
	HandleFSMMessage(bot, msgUpdate(chatID, "   "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого имени, получили: %s", cap.joined())
	}

	// Корректное имя
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingProjectName)
	mgr.GetData(chatID).EditingProjectID = pid
	HandleFSMMessage(bot, msgUpdate(chatID, "Новое имя"), mgr)
	if !cap.contains("обновлено") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	p, _ := database.GetProjectByID(pid)
	if p.Name != "Новое имя" {
		t.Errorf("имя не обновилось в БД: %q", p.Name)
	}
}

func TestFSMEditingProjectNameMissingProject(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(2007)
	mgr.SetState(chatID, fsm.StateEditingProjectName)
	mgr.GetData(chatID).EditingProjectID = 999999

	HandleFSMMessage(bot, msgUpdate(chatID, "Имя"), mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалось 'не найден', получили: %s", cap.joined())
	}
}

func TestFSMEditingProjectBudget(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2008)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "Не назначен")

	// Невалидный бюджет
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingProjectBudget)
	mgr.GetData(chatID).EditingProjectID = pid
	HandleFSMMessage(bot, msgUpdate(chatID, "abc"), mgr)
	if !cap.contains("числом") {
		t.Errorf("ожидалась ошибка валидации, получили: %s", cap.joined())
	}

	// Валидный бюджет
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingProjectBudget)
	mgr.GetData(chatID).EditingProjectID = pid
	HandleFSMMessage(bot, msgUpdate(chatID, "9999"), mgr)
	if !cap.contains("Бюджет обновлён") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	p, _ := database.GetProjectByID(pid)
	if p.Budget != 9999 {
		t.Errorf("бюджет не обновился: %v", p.Budget)
	}
}

func TestFSMEditingProjectBudgetMissingProject(t *testing.T) {
	bot, cap := newTestBot(t)
	mgr := fsm.NewManager()
	chatID := int64(2009)
	mgr.SetState(chatID, fsm.StateEditingProjectBudget)
	mgr.GetData(chatID).EditingProjectID = 999999

	HandleFSMMessage(bot, msgUpdate(chatID, "100"), mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалось 'не найден', получили: %s", cap.joined())
	}
}

func TestFSMCreatingTaskNameAndDescription(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2010)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "Не назначен")

	// Пустое имя задачи
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateCreatingTaskName)
	mgr.GetData(chatID).TaskProjectID = pid
	HandleFSMMessage(bot, msgUpdate(chatID, "  "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого имени, получили: %s", cap.joined())
	}

	// Корректное имя → переход к описанию
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "Задача А"), mgr)
	if mgr.GetState(chatID) != fsm.StateCreatingTaskDescription {
		t.Errorf("ожидался переход к описанию, состояние=%s", mgr.GetState(chatID))
	}

	// Описание /skip → задача создаётся
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "/skip"), mgr)
	if !cap.contains("Задача создана") {
		t.Errorf("ожидалось создание задачи, получили: %s", cap.joined())
	}
}

func TestFSMCreatingTaskWithDescription(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2011)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "Не назначен")

	mgr.SetState(chatID, fsm.StateCreatingTaskDescription)
	d := mgr.GetData(chatID)
	d.TaskProjectID = pid
	d.TaskName = "Задача Б"

	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "Полезное описание"), mgr)
	if !cap.contains("Задача создана") {
		t.Errorf("ожидалось создание задачи, получили: %s", cap.joined())
	}
}

func TestFSMEditingMasterName(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(2012)

	masters, err := database.GetAllMasters()
	if err != nil || len(masters) < 2 {
		t.Fatalf("нужны сидированные мастера: %v", err)
	}

	// Пустое ФИО
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingMasterName)
	mgr.GetData(chatID).EditingMasterID = masters[0].ID
	HandleFSMMessage(bot, msgUpdate(chatID, "  "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого ФИО, получили: %s", cap.joined())
	}

	// Дубликат (имя второго мастера)
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingMasterName)
	mgr.GetData(chatID).EditingMasterID = masters[0].ID
	HandleFSMMessage(bot, msgUpdate(chatID, masters[1].Name), mgr)
	if !cap.contains("Не удалось") {
		t.Errorf("ожидалась ошибка дубликата, получили: %s", cap.joined())
	}

	// Корректное новое ФИО
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingMasterName)
	mgr.GetData(chatID).EditingMasterID = masters[0].ID
	HandleFSMMessage(bot, msgUpdate(chatID, "Обновлённый Мастер"), mgr)
	if !cap.contains("ФИО обновлено") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	m, _ := database.GetMasterByID(masters[0].ID)
	if m.Name != "Обновлённый Мастер" {
		t.Errorf("ФИО не обновилось в БД: %q", m.Name)
	}
}

func TestStartHelpers(t *testing.T) {
	mgr := fsm.NewManager()

	bot, cap := newTestBot(t)
	StartProjectCreation(bot, 3001, mgr)
	if mgr.GetState(3001) != fsm.StateCreatingProjectType {
		t.Error("StartProjectCreation не установил состояние")
	}
	if cap.sends() == 0 {
		t.Error("StartProjectCreation не отправил сообщение")
	}

	bot, cap = newTestBot(t)
	StartTaskCreation(bot, 3002, mgr)
	if mgr.GetState(3002) != fsm.StateCreatingTask {
		t.Error("StartTaskCreation не установил состояние")
	}
	if cap.sends() == 0 {
		t.Error("StartTaskCreation не отправил сообщение")
	}
}
