package database

import (
	"testing"
	"time"
)

func TestParseDeadline(t *testing.T) {
	if d, err := ParseDeadline(""); err != nil || d != nil {
		t.Errorf("пустая строка: ожидалось nil,nil; получили %v,%v", d, err)
	}
	for _, s := range []string{"25.12.2026", "2026-12-25", "  25.12.2026 "} {
		d, err := ParseDeadline(s)
		if err != nil || d == nil {
			t.Errorf("ParseDeadline(%q) → %v,%v", s, d, err)
			continue
		}
		if d.Year() != 2026 || d.Month() != time.December || d.Day() != 25 {
			t.Errorf("ParseDeadline(%q) дал неверную дату: %v", s, d)
		}
	}
	if _, err := ParseDeadline("не дата"); err == nil {
		t.Error("ожидалась ошибка для мусора")
	}
}

func TestDeadlineRoundTrip(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(600, "Монтаж", "P", 1, "Не назначен")
	dl := time.Date(2027, 3, 15, 0, 0, 0, 0, time.UTC)

	tid, err := CreateTask(pid, "Со сроком", "", &dl)
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	task, err := GetTaskByID(tid)
	if err != nil {
		t.Fatalf("GetTaskByID: %v", err)
	}
	if task.Deadline == nil {
		t.Fatal("дедлайн не сохранился (round-trip сломан)")
	}
	if task.Deadline.Format("02.01.2006") != "15.03.2027" {
		t.Errorf("дедлайн round-trip = %v, want 15.03.2027", task.Deadline)
	}

	// Сброс срока
	if err := UpdateTaskDeadline(tid, nil); err != nil {
		t.Fatalf("UpdateTaskDeadline(nil): %v", err)
	}
	task, _ = GetTaskByID(tid)
	if task.Deadline != nil {
		t.Error("дедлайн должен быть сброшен")
	}
}

func TestIsTaskOverdue(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour)
	future := time.Now().Add(48 * time.Hour)

	if !IsTaskOverdue(Task{Status: "pending", Deadline: &past}) {
		t.Error("задача с прошедшим сроком должна быть просрочена")
	}
	if IsTaskOverdue(Task{Status: "completed", Deadline: &past}) {
		t.Error("выполненная задача не просрочена")
	}
	if IsTaskOverdue(Task{Status: "pending", Deadline: &future}) {
		t.Error("задача с будущим сроком не просрочена")
	}
	if IsTaskOverdue(Task{Status: "pending", Deadline: nil}) {
		t.Error("задача без срока не просрочена")
	}
}

func TestGetPortfolioStats(t *testing.T) {
	defer setupTestDB(t)()

	userID := int64(601)

	// Пустой портфель
	s, err := GetPortfolioStats(userID)
	if err != nil {
		t.Fatalf("GetPortfolioStats(пусто): %v", err)
	}
	if s.Projects != 0 || s.Tasks != 0 || s.Progress != 0 {
		t.Errorf("пустой портфель: %+v", s)
	}

	// Проект с 4 задачами: 1 выполнена, 1 просрочена
	pid, _ := CreateProject(userID, "Монтаж", "P", 5000, "Не назначен")
	ids := make([]int64, 0, 4)
	for i := 0; i < 4; i++ {
		id, _ := CreateTask(pid, "T", "", nil)
		ids = append(ids, id)
	}
	UpdateTaskStatus(ids[0], "completed")
	past := time.Now().Add(-72 * time.Hour)
	UpdateTaskDeadline(ids[1], &past) // просроченная (статус pending)

	s, err = GetPortfolioStats(userID)
	if err != nil {
		t.Fatalf("GetPortfolioStats: %v", err)
	}
	if s.Projects != 1 {
		t.Errorf("Projects = %d, want 1", s.Projects)
	}
	if s.TotalBudget != 5000 {
		t.Errorf("TotalBudget = %v, want 5000", s.TotalBudget)
	}
	if s.Tasks != 4 || s.Completed != 1 {
		t.Errorf("Tasks=%d Completed=%d, want 4/1", s.Tasks, s.Completed)
	}
	if s.Overdue != 1 {
		t.Errorf("Overdue = %d, want 1", s.Overdue)
	}
	if s.Progress != 25 {
		t.Errorf("Progress = %d, want 25", s.Progress)
	}
}

func TestUserSettings(t *testing.T) {
	defer setupTestDB(t)()

	userID := int64(602)

	// Первое чтение создаёт настройки по умолчанию
	s, err := GetUserSettings(userID)
	if err != nil {
		t.Fatalf("GetUserSettings: %v", err)
	}
	if !s.Notifications {
		t.Error("по умолчанию уведомления должны быть включены")
	}

	// Повторное чтение возвращает существующую запись
	if _, err := GetUserSettings(userID); err != nil {
		t.Fatalf("GetUserSettings(повтор): %v", err)
	}

	// Выключаем уведомления
	if err := SetNotifications(userID, false); err != nil {
		t.Fatalf("SetNotifications: %v", err)
	}
	s, _ = GetUserSettings(userID)
	if s.Notifications {
		t.Error("уведомления должны быть выключены")
	}

	// Включаем обратно
	SetNotifications(userID, true)
	s, _ = GetUserSettings(userID)
	if !s.Notifications {
		t.Error("уведомления должны быть включены")
	}
}

func TestUserSettingsErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetUserSettings(1); err == nil {
		t.Error("GetUserSettings на закрытой БД должен вернуть ошибку")
	}
	if err := SetNotifications(1, true); err == nil {
		t.Error("SetNotifications на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetPortfolioStats(1); err == nil {
		t.Error("GetPortfolioStats на закрытой БД должен вернуть ошибку")
	}
}
