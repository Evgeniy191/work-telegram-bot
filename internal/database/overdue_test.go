package database

import (
	"testing"
	"time"
)

func TestCountProjectOverdue(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(700, "Монтаж", "P", 100, "Не назначен")

	// Нет задач → 0
	n, err := CountProjectOverdue(pid)
	if err != nil {
		t.Fatalf("CountProjectOverdue: %v", err)
	}
	if n != 0 {
		t.Errorf("ожидалось 0 просрочек, получили %d", n)
	}

	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	t1, _ := CreateTask(pid, "Просрочена", "", &past)
	CreateTask(pid, "В сроке", "", &future)
	CreateTask(pid, "Без срока", "", nil)
	// Просроченная, но выполненная — не считается
	t4, _ := CreateTask(pid, "Выполнена", "", &past)
	UpdateTaskStatus(t4, "completed")
	_ = t1

	n, err = CountProjectOverdue(pid)
	if err != nil {
		t.Fatalf("CountProjectOverdue: %v", err)
	}
	if n != 1 {
		t.Errorf("ожидалась 1 просрочка, получили %d", n)
	}
}

func TestCountProjectOverdueError(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := CountProjectOverdue(1); err == nil {
		t.Error("ожидалась ошибка на закрытой БД")
	}
}
