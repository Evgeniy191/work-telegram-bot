package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func cb(t *testing.T, chatID int64, data string, mgr *fsm.Manager) *capture {
	t.Helper()
	bot, cap := newTestBot(t)
	HandleCallback(bot, cbUpdate(chatID, data), mgr)
	return cap
}

func TestCallbackStaticCases(t *testing.T) {
	mgr := fsm.NewManager()
	statics := []string{
		"project_1", "project_2", "project_3",
		"task_docs", "task_materials", "task_report",
		"back_projects", "confirm_yes", "confirm_no",
		"back_to_type", "back_to_name", "back_to_budget",
		"edit_type", "edit_name", "edit_budget", "edit_master",
		"save_project", "back_menu", "back_to_projects",
	}
	for _, data := range statics {
		t.Run(data, func(t *testing.T) {
			cap := cb(t, 1, data, mgr)
			if cap.sends() == 0 {
				t.Errorf("callback %q не отправил ответа", data)
			}
		})
	}
}

func TestCallbackProjectNewStartsCreation(t *testing.T) {
	mgr := fsm.NewManager()
	cap := cb(t, 4001, "project_new", mgr)
	if mgr.GetState(4001) != fsm.StateCreatingProjectType {
		t.Error("project_new должен запускать создание проекта")
	}
	if cap.sends() == 0 {
		t.Error("нет ответа на project_new")
	}
}

func TestCallbackProjectTypes(t *testing.T) {
	mgr := fsm.NewManager()
	types := map[string]string{
		"type_montazh":      "Монтаж",
		"type_remont":       "Ремонт",
		"type_ustanovka":    "Установка",
		"type_stroitelstvo": "Строительство",
	}
	for data, name := range types {
		t.Run(data, func(t *testing.T) {
			chatID := int64(4100 + len(data))
			cap := cb(t, chatID, data, mgr)
			if mgr.GetData(chatID).ProjectType != name {
				t.Errorf("тип %q не сохранён (got %q)", name, mgr.GetData(chatID).ProjectType)
			}
			if mgr.GetState(chatID) != fsm.StateCreatingProject {
				t.Error("ожидался переход к вводу названия")
			}
			if !cap.contains("Шаг 2") {
				t.Errorf("ожидался шаг 2, получили: %s", cap.joined())
			}
		})
	}
}

func TestCallbackMasterPickCreatesProject(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4200)

	masters, _ := database.GetAllMasters()
	master := masters[0]

	// Готовим состояние как после ввода бюджета
	mgr.SetState(chatID, fsm.StateCreatingProjectMaster)
	d := mgr.GetData(chatID)
	d.ProjectType = "Монтаж"
	d.ProjectName = "Проект через выбор мастера"
	d.ProjectBudget = "1000.00"

	cap := cb(t, chatID, fmt.Sprintf("master_pick_%d", master.ID), mgr)
	if !cap.contains("Проект создан") {
		t.Errorf("ожидалось создание проекта, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("состояние должно сброситься после создания")
	}

	// Проект реально в БД
	projects, _ := database.GetUserProjects(chatID)
	if len(projects) != 1 {
		t.Fatalf("ожидался 1 проект в БД, получили %d", len(projects))
	}
	if database.GetMasterNameByID(projects[0].MasterID) != master.Name {
		t.Error("мастер не сохранён в проекте")
	}
}

func TestCallbackMasterNone(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4201)
	mgr.SetState(chatID, fsm.StateCreatingProjectMaster)
	d := mgr.GetData(chatID)
	d.ProjectType = "Ремонт"
	d.ProjectName = "Без мастера"
	d.ProjectBudget = "500.00"

	cap := cb(t, chatID, "master_none", mgr)
	if !cap.contains("Проект создан") {
		t.Errorf("ожидалось создание проекта, получили: %s", cap.joined())
	}
}

func TestCallbackMasterPickUpdatesExistingProject(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4202)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "Не назначен")

	masters, _ := database.GetAllMasters()
	master := masters[1]

	// Контекст смены мастера у существующего проекта
	mgr.GetData(chatID).EditingProjectID = pid

	cap := cb(t, chatID, fmt.Sprintf("master_pick_%d", master.ID), mgr)
	if !cap.contains("Ответственный обновлён") {
		t.Errorf("ожидалось обновление мастера, получили: %s", cap.joined())
	}
	p, _ := database.GetProjectByID(pid)
	if database.GetMasterNameByID(p.MasterID) != master.Name {
		t.Error("мастер не обновился в БД")
	}
	if mgr.GetData(chatID).EditingProjectID != 0 {
		t.Error("EditingProjectID должен сброситься")
	}
}

func TestCallbackMasterPickInMemory(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4203)
	masters, _ := database.GetAllMasters()

	// Нет состояния создания и нет EditingProjectID → in-memory выбор
	cap := cb(t, chatID, fmt.Sprintf("master_pick_%d", masters[0].ID), mgr)
	if !cap.contains("Ответственный выбран") {
		t.Errorf("ожидался in-memory выбор, получили: %s", cap.joined())
	}
}

func TestCallbackMasterPickErrors(t *testing.T) {
	mgr := fsm.NewManager()

	cap := cb(t, 4204, "master_pick_notanumber", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка неверного ID, получили: %s", cap.joined())
	}

	cap = cb(t, 4205, "master_pick_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackEditLastProject(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4300)

	// Без данных
	cap := cb(t, chatID, "edit_last_project", mgr)
	if !cap.contains("Нет данных") {
		t.Errorf("ожидалось 'Нет данных', получили: %s", cap.joined())
	}

	// С данными
	d := mgr.GetData(chatID)
	d.ProjectType = "Монтаж"
	d.ProjectName = "Проект"
	d.ProjectBudget = "1000.00"
	d.ProjectMaster = "Иванов Иван"
	cap = cb(t, chatID, "edit_last_project", mgr)
	if !cap.contains("Редактирование проекта") {
		t.Errorf("ожидалось меню редактирования, получили: %s", cap.joined())
	}
}

func TestCallbackEditByID(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4400)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "Иванов Иван")

	// edit_name_<id>
	cap := cb(t, chatID, fmt.Sprintf("edit_name_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingProjectName {
		t.Error("edit_name_<id> не установил состояние")
	}
	_ = cap

	// edit_budget_<id>
	cb(t, chatID, fmt.Sprintf("edit_budget_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingProjectBudget {
		t.Error("edit_budget_<id> не установил состояние")
	}

	// edit_master_<id> (смена мастера у проекта)
	cap = cb(t, chatID, fmt.Sprintf("edit_master_%d", pid), mgr)
	if mgr.GetData(chatID).EditingProjectID != pid {
		t.Error("edit_master_<id> не сохранил EditingProjectID")
	}
	if cap.sends() == 0 {
		t.Error("edit_master_<id> не отправил клавиатуру")
	}
}

func TestCallbackEditMasterName(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4500)
	masters, _ := database.GetAllMasters()

	cap := cb(t, chatID, fmt.Sprintf("edit_master_name_%d", masters[0].ID), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingMasterName {
		t.Error("edit_master_name_<id> не установил состояние")
	}
	if mgr.GetData(chatID).EditingMasterID != masters[0].ID {
		t.Error("EditingMasterID не сохранён")
	}
	if !cap.contains("Изменение ФИО") {
		t.Errorf("ожидался запрос ФИО, получили: %s", cap.joined())
	}

	// Ошибки
	cap = cb(t, chatID, "edit_master_name_xx", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "edit_master_name_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackEditProjectView(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4600)
	pid, _ := database.CreateProject(chatID, "Монтаж", "Проект", 100, "Иванов Иван")

	cap := cb(t, chatID, fmt.Sprintf("edit_project_%d", pid), mgr)
	if !cap.contains("Редактирование проекта") {
		t.Errorf("ожидалось меню, получили: %s", cap.joined())
	}

	// Неверный ID
	cap = cb(t, chatID, "edit_project_abc", mgr)
	if !cap.contains("неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующий проект
	cap = cb(t, chatID, "edit_project_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackDeleteProjectFlow(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4700)
	pid, _ := database.CreateProject(chatID, "Монтаж", "Удаляемый", 100, "")

	// Запрос подтверждения
	cap := cb(t, chatID, fmt.Sprintf("delete_project_%d", pid), mgr)
	if !cap.contains("Подтверждение удаления") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}

	// Неверный ID
	cap = cb(t, chatID, "delete_project_xx", mgr)
	if !cap.contains("неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующий
	cap = cb(t, chatID, "delete_project_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}

	// Подтверждённое удаление
	cap = cb(t, chatID, fmt.Sprintf("confirm_delete_%d", pid), mgr)
	if !cap.contains("Проект удалён") {
		t.Errorf("ожидалось удаление, получили: %s", cap.joined())
	}
	if _, err := database.GetProjectByID(pid); err == nil {
		t.Error("проект не удалён из БД")
	}

	// confirm_delete с неверным ID
	cap = cb(t, chatID, "confirm_delete_xx", mgr)
	if !cap.contains("неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
}

func TestCallbackViewTasks(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4800)

	// Проект без задач
	emptyPID, _ := database.CreateProject(chatID, "Монтаж", "Пустой", 100, "")
	cap := cb(t, chatID, fmt.Sprintf("view_tasks_%d", emptyPID), mgr)
	if !cap.contains("нет задач") {
		t.Errorf("ожидалось 'нет задач', получили: %s", cap.joined())
	}

	// Проект с задачами (разные статусы → разные ветки кнопок)
	pid, _ := dbCreateProjectWithTasks(chatID, 3, 1)
	tasks, _ := database.GetProjectTasks(pid)
	database.UpdateTaskStatus(tasks[1].ID, "in_progress")

	cap = cb(t, chatID, fmt.Sprintf("view_tasks_%d", pid), mgr)
	if !cap.contains("Прогресс проекта") {
		t.Errorf("ожидался прогресс, получили: %s", cap.joined())
	}

	// Неверный ID
	cap = cb(t, chatID, "view_tasks_xx", mgr)
	if !cap.contains("неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующий проект
	cap = cb(t, chatID, "view_tasks_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackAddTask(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(4900)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")

	cap := cb(t, chatID, fmt.Sprintf("add_task_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateCreatingTaskName {
		t.Error("add_task_<id> не установил состояние")
	}
	if mgr.GetData(chatID).TaskProjectID != pid {
		t.Error("TaskProjectID не сохранён")
	}
	_ = cap

	cap = cb(t, chatID, "add_task_xx", mgr)
	if !cap.contains("неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
}

func TestCallbackTaskStatus(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(5000)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)

	cap := cb(t, chatID, fmt.Sprintf("task_status_%d_in_progress", tid), mgr)
	if !cap.contains("Статус изменён") {
		t.Errorf("ожидалось изменение статуса, получили: %s", cap.joined())
	}
	task, _ := database.GetTaskByID(tid)
	if task.Status != "in_progress" {
		t.Errorf("статус не обновлён: %q", task.Status)
	}

	// Плохой формат
	cap = cb(t, chatID, "task_status_5", mgr)
	if !cap.contains("формат") {
		t.Errorf("ожидалась ошибка формата, получили: %s", cap.joined())
	}

	// Плохой ID
	cap = cb(t, chatID, "task_status_abc_pending", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующая задача
	cap = cb(t, chatID, "task_status_999999_pending", mgr)
	if !cap.contains("Ошибка") {
		t.Errorf("ожидалась ошибка, получили: %s", cap.joined())
	}
}

func TestCallbackDeleteTask(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(5100)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)

	cap := cb(t, chatID, fmt.Sprintf("delete_task_%d", tid), mgr)
	if !cap.contains("Задача удалена") {
		t.Errorf("ожидалось удаление, получили: %s", cap.joined())
	}

	// Несуществующая
	cap = cb(t, chatID, "delete_task_999999", mgr)
	if !cap.contains("Ошибка удаления") {
		t.Errorf("ожидалась ошибка удаления, получили: %s", cap.joined())
	}
}

func TestCallbackUnknown(t *testing.T) {
	mgr := fsm.NewManager()
	cap := cb(t, 5200, "совершенно_неизвестный", mgr)
	if !cap.contains("Неизвестная команда") {
		t.Errorf("ожидалось 'Неизвестная команда', получили: %s", cap.joined())
	}
}

// --- Юнит-тесты вспомогательных функций ---

func TestGetProjectTypeEmoji(t *testing.T) {
	cases := map[string]string{
		"Монтаж":        "🔧",
		"Ремонт":        "🛠️",
		"Установка":     "⚙️",
		"Строительство": "🏗️",
		"Другое":        "📋",
	}
	for in, want := range cases {
		if got := getProjectTypeEmoji(in); got != want {
			t.Errorf("getProjectTypeEmoji(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	cases := map[float64]string{
		0:        "0",
		1000:     "1 000",
		1234567:  "1 234 567",
		12500.50: "12 500.50",
		999.99:   "999.99",
	}
	for in, want := range cases {
		if got := formatMoney(in); got != want {
			t.Errorf("formatMoney(%v)=%q, want %q", in, got, want)
		}
	}
}

func TestProgressBar(t *testing.T) {
	if got := progressBar(0); got != strings.Repeat("▱", 10) {
		t.Errorf("progressBar(0)=%q", got)
	}
	if got := progressBar(100); got != strings.Repeat("▰", 10) {
		t.Errorf("progressBar(100)=%q", got)
	}
	// Клемпинг
	if got := progressBar(-5); got != strings.Repeat("▱", 10) {
		t.Errorf("progressBar(-5) должен клемпиться к 0: %q", got)
	}
	if got := progressBar(150); got != strings.Repeat("▰", 10) {
		t.Errorf("progressBar(150) должен клемпиться к 100: %q", got)
	}
	// 30% → 3 заполненных
	if got := progressBar(30); !strings.HasPrefix(got, strings.Repeat("▰", 3)) {
		t.Errorf("progressBar(30)=%q", got)
	}
}
