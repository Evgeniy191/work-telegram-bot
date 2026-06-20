package database

import "testing"

func TestExpensesAndBudgetStatus(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1000, "Монтаж", "P", 10000, "Не назначен")

	// Пустой бюджет-статус
	st, err := GetProjectBudgetStatus(pid)
	if err != nil {
		t.Fatalf("GetProjectBudgetStatus: %v", err)
	}
	if st.Budget != 10000 || st.Spent != 0 || st.Remaining != 10000 || st.Percent != 0 || st.Over {
		t.Errorf("пустой статус неверен: %+v", st)
	}

	// Добавляем расходы
	if _, err := AddExpense(pid, 3000, "материалы"); err != nil {
		t.Fatalf("AddExpense: %v", err)
	}
	AddExpense(pid, 2000, "")

	spent, _ := GetProjectSpent(pid)
	if spent != 5000 {
		t.Errorf("GetProjectSpent = %v, want 5000", spent)
	}

	st, _ = GetProjectBudgetStatus(pid)
	if st.Spent != 5000 || st.Remaining != 5000 || st.Percent != 50 || st.Over {
		t.Errorf("статус после расходов неверен: %+v", st)
	}

	exps, _ := GetProjectExpenses(pid)
	if len(exps) != 2 {
		t.Errorf("ожидалось 2 расхода, получили %d", len(exps))
	}

	// Перерасход
	AddExpense(pid, 8000, "доп. работы")
	st, _ = GetProjectBudgetStatus(pid)
	if !st.Over {
		t.Error("ожидался признак перерасхода")
	}
	if st.Remaining >= 0 {
		t.Errorf("остаток при перерасходе должен быть отрицательным: %v", st.Remaining)
	}
}

func TestAddExpenseValidation(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1001, "Монтаж", "P", 100, "Не назначен")

	if _, err := AddExpense(pid, -5, "минус"); err == nil {
		t.Error("ожидалась ошибка для отрицательной суммы")
	}
	if _, err := AddExpense(999999, 100, "нет проекта"); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта")
	}
}

func TestBudgetZeroBudget(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1002, "Монтаж", "P", 0, "Не назначен")
	AddExpense(pid, 500, "")

	st, _ := GetProjectBudgetStatus(pid)
	if st.Percent != 0 {
		t.Errorf("при нулевом бюджете процент = %d, ожидался 0 (без деления на ноль)", st.Percent)
	}
	if !st.Over {
		t.Error("расход при нулевом бюджете — это перерасход")
	}
}

func TestExpensesErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetProjectExpenses(1); err == nil {
		t.Error("GetProjectExpenses на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetProjectSpent(1); err == nil {
		t.Error("GetProjectSpent на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetProjectBudgetStatus(1); err == nil {
		t.Error("GetProjectBudgetStatus на закрытой БД должен вернуть ошибку")
	}
}
