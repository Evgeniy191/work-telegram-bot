package database

import (
	"os"
	"testing"
)

// setupTestDB создаёт временную БД для тестов
func setupTestDB(t *testing.T) func() {
	t.Helper()

	tmp, err := os.MkdirTemp("", "bot-test-*")
	if err != nil {
		t.Fatalf("не удалось создать temp dir: %v", err)
	}

	// InitDB создаёт ./data/bot.db относительно рабочей директории.
	// После открытия соединения сразу возвращаем cwd: дескриптор БД остаётся
	// валидным, а тесты не выполняются с изменённой рабочей директорией
	// (важно для корректной записи coverage-профиля в режиме -coverpkg).
	cwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	os.Chdir(cwd)

	return func() {
		CloseDB()
		os.RemoveAll(tmp)
	}
}

func TestCalcProgress(t *testing.T) {
	cases := []struct {
		total, completed, want int
	}{
		{0, 0, 0},
		{10, 0, 0},
		{10, 3, 30},
		{10, 10, 100},
		{3, 1, 33},
		{3, 2, 67},
		{7, 7, 100},
	}
	for _, c := range cases {
		if got := CalcProgress(c.total, c.completed); got != c.want {
			t.Errorf("CalcProgress(%d,%d)=%d, want %d", c.total, c.completed, got, c.want)
		}
	}
}

func TestMastersEditFlow(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	masters, err := GetAllMasters()
	if err != nil {
		t.Fatalf("GetAllMasters: %v", err)
	}
	if len(masters) == 0 {
		t.Fatal("ожидались сидированные мастера, получили 0")
	}

	first := masters[0]

	// Переименование ФИО
	newName := "Тестов Тест Тестович"
	if err := UpdateMasterName(first.ID, newName); err != nil {
		t.Fatalf("UpdateMasterName: %v", err)
	}

	got, err := GetMasterByID(first.ID)
	if err != nil {
		t.Fatalf("GetMasterByID: %v", err)
	}
	if got.Name != newName {
		t.Errorf("ФИО не обновилось: got %q, want %q", got.Name, newName)
	}

	// Дубликат ФИО (имя второго мастера) должен отклоняться
	if len(masters) > 1 {
		if err := UpdateMasterName(first.ID, masters[1].Name); err == nil {
			t.Error("ожидалась ошибка при дубликате ФИО, получили nil")
		}
	}
}

func TestProjectProgressFromTasks(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	projectID, err := CreateProject(42, "Монтаж", "Тестовый проект", 1000, "Не назначен")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}

	// Нет задач → 0%
	percent, total, completed, err := GetProjectProgress(projectID)
	if err != nil {
		t.Fatalf("GetProjectProgress (пусто): %v", err)
	}
	if percent != 0 || total != 0 || completed != 0 {
		t.Errorf("пустой проект: percent=%d total=%d completed=%d, want 0/0/0", percent, total, completed)
	}

	// 4 задачи, 1 выполнена → 25%
	ids := make([]int64, 0, 4)
	for i := 0; i < 4; i++ {
		id, err := CreateTask(projectID, "Задача", "", nil)
		if err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
		ids = append(ids, id)
	}
	if err := UpdateTaskStatus(ids[0], "completed"); err != nil {
		t.Fatalf("UpdateTaskStatus: %v", err)
	}

	percent, total, completed, err = GetProjectProgress(projectID)
	if err != nil {
		t.Fatalf("GetProjectProgress: %v", err)
	}
	if total != 4 || completed != 1 || percent != 25 {
		t.Errorf("прогресс: percent=%d total=%d completed=%d, want 25/4/1", percent, total, completed)
	}
}
