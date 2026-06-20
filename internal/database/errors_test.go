package database

import (
	"testing"
	"time"
)

// TestUpdateProjectWithMaster покрывает ветку поиска мастера по имени в UpdateProject.
func TestUpdateProjectWithMaster(t *testing.T) {
	defer setupTestDB(t)()

	id, err := CreateProject(300, "Монтаж", "P", 1, "Не назначен")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Назначаем существующего мастера через UpdateProject
	if err := UpdateProject(id, "P", 2, "Петров Пётр"); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	p, _ := GetProjectByID(id)
	if GetMasterNameByID(p.MasterID) != "Петров Пётр" {
		t.Errorf("мастер не назначен через UpdateProject: %v", p.MasterID)
	}
}

// TestDatabaseErrorPaths закрывает соединение и проверяет, что функции
// корректно возвращают ошибки (покрытие ветвей обработки ошибок).
func TestDatabaseErrorPaths(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Принудительно закрываем соединение — все запросы начнут падать.
	if err := DB.Close(); err != nil {
		t.Fatalf("DB.Close: %v", err)
	}

	if _, err := GetAllMasters(); err == nil {
		t.Error("GetAllMasters должен вернуть ошибку на закрытой БД")
	}
	if _, err := GetMasterByID(1); err == nil {
		t.Error("GetMasterByID должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateMasterName(1, "X"); err == nil {
		t.Error("UpdateMasterName должен вернуть ошибку на закрытой БД")
	}
	if _, err := GetUserProjects(1); err == nil {
		t.Error("GetUserProjects должен вернуть ошибку на закрытой БД")
	}
	if _, err := GetProjectByID(1); err == nil {
		t.Error("GetProjectByID должен вернуть ошибку на закрытой БД")
	}
	if _, err := CreateProject(1, "T", "N", 1, "Иванов Иван"); err == nil {
		t.Error("CreateProject должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateProject(1, "N", 1, ""); err == nil {
		t.Error("UpdateProject должен вернуть ошибку на закрытой БД")
	}
	if err := DeleteProject(1); err == nil {
		t.Error("DeleteProject должен вернуть ошибку на закрытой БД")
	}
	if _, err := CreateTask(1, "N", "", nil); err == nil {
		t.Error("CreateTask должен вернуть ошибку на закрытой БД")
	}
	if _, err := GetProjectTasks(1); err == nil {
		t.Error("GetProjectTasks должен вернуть ошибку на закрытой БД")
	}
	if _, err := GetTaskByID(1); err == nil {
		t.Error("GetTaskByID должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateTaskStatus(1, "pending"); err == nil {
		t.Error("UpdateTaskStatus должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateTask(1, "N", "D"); err == nil {
		t.Error("UpdateTask должен вернуть ошибку на закрытой БД")
	}
	if err := DeleteTask(1); err == nil {
		t.Error("DeleteTask должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateTaskName(1, "N"); err == nil {
		t.Error("UpdateTaskName должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateTaskDescription(1, "D"); err == nil {
		t.Error("UpdateTaskDescription должен вернуть ошибку на закрытой БД")
	}
	dl := time.Now()
	if err := UpdateTaskDeadline(1, &dl); err == nil {
		t.Error("UpdateTaskDeadline(set) должен вернуть ошибку на закрытой БД")
	}
	if err := UpdateTaskDeadline(1, nil); err == nil {
		t.Error("UpdateTaskDeadline(nil) должен вернуть ошибку на закрытой БД")
	}
	if _, _, _, _, err := CountProjectTasks(1); err == nil {
		t.Error("CountProjectTasks должен вернуть ошибку на закрытой БД")
	}
	if _, _, _, err := GetProjectProgress(1); err == nil {
		t.Error("GetProjectProgress должен вернуть ошибку на закрытой БД")
	}
}
