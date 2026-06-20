package database

import (
	"fmt"
	"log"
)

// BudgetStatus — план/факт по бюджету проекта.
type BudgetStatus struct {
	Budget    float64 // плановый бюджет
	Spent     float64 // освоено (сумма расходов)
	Remaining float64 // остаток (может быть отрицательным при перерасходе)
	Percent   int     // % освоения
	Over      bool    // признак перерасхода
}

// AddExpense — добавляет расход к проекту.
func AddExpense(projectID int64, amount float64, description string) (int64, error) {
	if amount < 0 {
		return 0, fmt.Errorf("сумма расхода не может быть отрицательной")
	}

	var exists bool
	if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", projectID).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("проект ID=%d не найден", projectID)
	}

	result, err := DB.Exec(`
		INSERT INTO expenses (project_id, amount, description, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`, projectID, amount, description)
	if err != nil {
		log.Printf("❌ Ошибка добавления расхода: %v", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	log.Printf("💸 Расход ID=%d (%.2f) добавлен к проекту %d", id, amount, projectID)
	return id, nil
}

// GetProjectExpenses — список расходов проекта (новые первыми).
func GetProjectExpenses(projectID int64) ([]Expense, error) {
	rows, err := DB.Query(`
		SELECT id, project_id, amount, IFNULL(description, ''), created_at
		FROM expenses WHERE project_id = ?
		ORDER BY created_at DESC, id DESC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var e Expense
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Amount, &e.Description, &e.CreatedAt); err != nil {
			log.Printf("❌ Ошибка чтения расхода: %v", err)
			continue
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

// GetProjectSpent — сумма расходов по проекту.
func GetProjectSpent(projectID int64) (float64, error) {
	var spent float64
	err := DB.QueryRow(
		"SELECT IFNULL(SUM(amount), 0) FROM expenses WHERE project_id = ?",
		projectID,
	).Scan(&spent)
	return spent, err
}

// GetProjectBudgetStatus — план/факт по бюджету проекта.
func GetProjectBudgetStatus(projectID int64) (BudgetStatus, error) {
	var s BudgetStatus

	p, err := GetProjectByID(projectID)
	if err != nil {
		return s, err
	}
	spent, err := GetProjectSpent(projectID)
	if err != nil {
		return s, err
	}

	s.Budget = p.Budget
	s.Spent = spent
	s.Remaining = p.Budget - spent
	s.Over = spent > p.Budget
	if p.Budget > 0 {
		s.Percent = int(spent/p.Budget*100 + 0.5)
	}
	return s, nil
}
