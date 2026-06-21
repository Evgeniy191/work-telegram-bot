package database

import (
	"testing"
	"time"
)

func TestGetPortfolioDashboard(t *testing.T) {
	defer setupTestDB(t)()

	userID := int64(900)

	// Пустой портфель.
	d, err := GetPortfolioDashboard(userID)
	if err != nil {
		t.Fatalf("GetPortfolioDashboard(пусто): %v", err)
	}
	if d.Projects != 0 || d.Spent != 0 || len(d.StatusCounts) != 0 {
		t.Errorf("пустой портфель: %+v", d)
	}

	// Три проекта: два «В работе», один «Завершён».
	p1, _ := CreateProject(userID, "Монтаж", "P1", 10000, "Не назначен")
	CreateProject(userID, "Ремонт", "P2", 5000, "Не назначен")
	p3, _ := CreateProject(userID, "Установка", "P3", 0, "Не назначен")
	if _, err := DB.Exec("UPDATE projects SET status = 'Завершён' WHERE id = ?", p3); err != nil {
		t.Fatalf("смена статуса проекта: %v", err)
	}

	// Освоение бюджета по p1: 3000 + 1000 = 4000 из 15000 (≈27%).
	AddExpense(p1, 3000, "материалы")
	AddExpense(p1, 1000, "работа")

	// Задачи p1: одна выполнена, одна просрочена.
	tDone, _ := CreateTask(p1, "T1", "", nil)
	UpdateTaskStatus(tDone, "completed")
	tOver, _ := CreateTask(p1, "T2", "", nil)
	past := time.Now().Add(-48 * time.Hour)
	UpdateTaskDeadline(tOver, &past)

	d, err = GetPortfolioDashboard(userID)
	if err != nil {
		t.Fatalf("GetPortfolioDashboard: %v", err)
	}

	if d.Projects != 3 {
		t.Errorf("Projects = %d, want 3", d.Projects)
	}
	if d.TotalBudget != 15000 {
		t.Errorf("TotalBudget = %v, want 15000", d.TotalBudget)
	}
	if d.Spent != 4000 {
		t.Errorf("Spent = %v, want 4000", d.Spent)
	}
	if d.Remaining != 11000 {
		t.Errorf("Remaining = %v, want 11000", d.Remaining)
	}
	if d.BudgetPercent != 27 { // 4000/15000 = 26.67 → 27
		t.Errorf("BudgetPercent = %d, want 27", d.BudgetPercent)
	}
	if d.OverBudget {
		t.Error("OverBudget = true, want false")
	}
	if d.Tasks != 2 || d.Completed != 1 || d.Overdue != 1 {
		t.Errorf("задачи: Tasks=%d Completed=%d Overdue=%d, want 2/1/1", d.Tasks, d.Completed, d.Overdue)
	}

	// Разбивка по статусам: «В работе» (2) первым, «Завершён» (1) вторым.
	if len(d.StatusCounts) != 2 {
		t.Fatalf("StatusCounts: %+v, want 2 статуса", d.StatusCounts)
	}
	if d.StatusCounts[0].Status != "В работе" || d.StatusCounts[0].Count != 2 {
		t.Errorf("StatusCounts[0] = %+v, want {В работе 2}", d.StatusCounts[0])
	}
	if d.StatusCounts[1].Status != "Завершён" || d.StatusCounts[1].Count != 1 {
		t.Errorf("StatusCounts[1] = %+v, want {Завершён 1}", d.StatusCounts[1])
	}
}

func TestGetPortfolioDashboardOverBudget(t *testing.T) {
	defer setupTestDB(t)()

	userID := int64(901)
	pid, _ := CreateProject(userID, "Монтаж", "P", 1000, "Не назначен")
	AddExpense(pid, 1500, "перерасход")

	d, err := GetPortfolioDashboard(userID)
	if err != nil {
		t.Fatalf("GetPortfolioDashboard: %v", err)
	}
	if !d.OverBudget {
		t.Error("OverBudget = false, want true (1500 > 1000)")
	}
	if d.Remaining != -500 {
		t.Errorf("Remaining = %v, want -500", d.Remaining)
	}
	if d.BudgetPercent != 150 {
		t.Errorf("BudgetPercent = %d, want 150", d.BudgetPercent)
	}
}

func TestGetPortfolioDashboardErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetPortfolioDashboard(1); err == nil {
		t.Error("GetPortfolioDashboard на закрытой БД должен вернуть ошибку")
	}
}
