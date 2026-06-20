package database

import (
	"testing"
	"time"
)

func TestProjectCRUD(t *testing.T) {
	defer setupTestDB(t)()

	// Create с известным мастером (есть в сидах)
	id, err := CreateProject(100, "Монтаж", "Труба А", 5000, "Иванов Иван")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if id == 0 {
		t.Fatal("ожидался ненулевой ID проекта")
	}

	// GetByID
	p, err := GetProjectByID(id)
	if err != nil {
		t.Fatalf("GetProjectByID: %v", err)
	}
	if p.Name != "Труба А" || p.Type != "Монтаж" || p.Budget != 5000 {
		t.Errorf("неверные данные проекта: %+v", p)
	}
	if p.MasterID == nil {
		t.Error("ожидался назначенный мастер")
	}
	if name := GetMasterNameByID(p.MasterID); name != "Иванов Иван" {
		t.Errorf("GetMasterNameByID = %q, want Иванов Иван", name)
	}

	// Update — сменим имя, бюджет и мастера на "Не назначен"
	if err := UpdateProject(id, "Труба Б", 7500, "Не назначен"); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	p, _ = GetProjectByID(id)
	if p.Name != "Труба Б" || p.Budget != 7500 {
		t.Errorf("обновление не применилось: %+v", p)
	}
	if p.MasterID != nil {
		t.Errorf("мастер должен быть nil, получили %v", *p.MasterID)
	}
	if GetMasterNameByID(p.MasterID) != "Не назначен" {
		t.Error("GetMasterNameByID(nil) должно быть 'Не назначен'")
	}

	// GetUserProjects
	projects, err := GetUserProjects(100)
	if err != nil {
		t.Fatalf("GetUserProjects: %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("ожидался 1 проект, получили %d", len(projects))
	}

	// Delete
	if err := DeleteProject(id); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if _, err := GetProjectByID(id); err == nil {
		t.Error("ожидалась ошибка для удалённого проекта")
	}
}

func TestCreateProjectUnknownMaster(t *testing.T) {
	defer setupTestDB(t)()

	id, err := CreateProject(101, "Ремонт", "Проект", 100, "Несуществующий Мастер")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	p, _ := GetProjectByID(id)
	if p.MasterID != nil {
		t.Error("неизвестный мастер должен дать master_id = NULL")
	}
}

func TestGetUserProjectsEmpty(t *testing.T) {
	defer setupTestDB(t)()

	projects, err := GetUserProjects(999)
	if err != nil {
		t.Fatalf("GetUserProjects: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("ожидался пустой список, получили %d", len(projects))
	}
}

func TestDeleteProjectNotFound(t *testing.T) {
	defer setupTestDB(t)()

	if err := DeleteProject(123456); err == nil {
		t.Error("ожидалась ошибка при удалении несуществующего проекта")
	}
}

func TestGetMasterNameByIDInvalid(t *testing.T) {
	defer setupTestDB(t)()

	bad := int64(987654)
	if name := GetMasterNameByID(&bad); name != "Не назначен" {
		t.Errorf("для несуществующего ID ожидалось 'Не назначен', получили %q", name)
	}
}

func TestTaskCRUD(t *testing.T) {
	defer setupTestDB(t)()

	projectID, err := CreateProject(200, "Установка", "Проект задач", 1, "Не назначен")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Create
	taskID, err := CreateTask(projectID, "Задача 1", "Описание", nil)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	// GetByID
	task, err := GetTaskByID(taskID)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.Name != "Задача 1" || task.Description != "Описание" || task.Status != "pending" {
		t.Errorf("неверные данные задачи: %+v", task)
	}

	// GetProjectTasks
	tasks, err := GetProjectTasks(projectID)
	if err != nil {
		t.Fatalf("GetProjectTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("ожидалась 1 задача, получили %d", len(tasks))
	}

	// UpdateTask
	if err := UpdateTask(taskID, "Задача 1b", "Описание b"); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	task, _ = GetTaskByID(taskID)
	if task.Name != "Задача 1b" || task.Description != "Описание b" {
		t.Errorf("UpdateTask не применился: %+v", task)
	}

	// UpdateTaskName / UpdateTaskDescription
	if err := UpdateTaskName(taskID, "Имя2"); err != nil {
		t.Fatalf("UpdateTaskName: %v", err)
	}
	if err := UpdateTaskDescription(taskID, "Опис2"); err != nil {
		t.Fatalf("UpdateTaskDescription: %v", err)
	}
	task, _ = GetTaskByID(taskID)
	if task.Name != "Имя2" || task.Description != "Опис2" {
		t.Errorf("частичные обновления не применились: %+v", task)
	}

	// UpdateTaskStatus
	if err := UpdateTaskStatus(taskID, "completed"); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}
	task, _ = GetTaskByID(taskID)
	if task.Status != "completed" {
		t.Errorf("статус = %q, want completed", task.Status)
	}

	// UpdateTaskDeadline — установка и сброс
	dl := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := UpdateTaskDeadline(taskID, &dl); err != nil {
		t.Fatalf("UpdateTaskDeadline(set): %v", err)
	}
	if err := UpdateTaskDeadline(taskID, nil); err != nil {
		t.Fatalf("UpdateTaskDeadline(nil): %v", err)
	}

	// CountProjectTasks
	total, _, _, completed, err := CountProjectTasks(projectID)
	if err != nil {
		t.Fatalf("CountProjectTasks: %v", err)
	}
	if total != 1 || completed != 1 {
		t.Errorf("CountProjectTasks: total=%d completed=%d, want 1/1", total, completed)
	}

	// Delete
	if err := DeleteTask(taskID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if _, err := GetTaskByID(taskID); err == nil {
		t.Error("ожидалась ошибка для удалённой задачи")
	}
}

func TestCreateTaskInvalidProject(t *testing.T) {
	defer setupTestDB(t)()

	if _, err := CreateTask(424242, "x", "", nil); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта")
	}
}

func TestUpdateTaskStatusInvalid(t *testing.T) {
	defer setupTestDB(t)()

	projectID, _ := CreateProject(201, "Монтаж", "P", 1, "")
	taskID, _ := CreateTask(projectID, "T", "", nil)

	if err := UpdateTaskStatus(taskID, "не_статус"); err == nil {
		t.Error("ожидалась ошибка для неверного статуса")
	}
	if err := UpdateTaskStatus(424242, "pending"); err == nil {
		t.Error("ожидалась ошибка для несуществующей задачи")
	}
}

func TestDeleteTaskNotFound(t *testing.T) {
	defer setupTestDB(t)()
	if err := DeleteTask(424242); err == nil {
		t.Error("ожидалась ошибка при удалении несуществующей задачи")
	}
}

func TestTaskStatusHelpers(t *testing.T) {
	cases := []struct {
		status, emoji, name string
	}{
		{"pending", "🕐", "Ожидает"},
		{"in_progress", "⚙️", "В процессе"},
		{"completed", "✅", "Выполнена"},
		{"unknown", "❓", "Неизвестно"},
	}
	for _, c := range cases {
		if e := GetTaskStatusEmoji(c.status); e != c.emoji {
			t.Errorf("GetTaskStatusEmoji(%q)=%q, want %q", c.status, e, c.emoji)
		}
		if n := GetTaskStatusName(c.status); n != c.name {
			t.Errorf("GetTaskStatusName(%q)=%q, want %q", c.status, n, c.name)
		}
	}
}

func TestMasterErrors(t *testing.T) {
	defer setupTestDB(t)()

	if _, err := GetMasterByID(424242); err == nil {
		t.Error("ожидалась ошибка для несуществующего мастера")
	}
	if err := UpdateMasterName(424242, "Кто-то"); err == nil {
		t.Error("ожидалась ошибка при обновлении несуществующего мастера")
	}
}

func TestCloseDBNil(t *testing.T) {
	old := DB
	DB = nil
	if err := CloseDB(); err != nil {
		t.Errorf("CloseDB() с nil должно вернуть nil, получили %v", err)
	}
	DB = old
}
