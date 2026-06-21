package database

import "testing"

func TestStagesCRUDAndProgress(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1200, "Строительство", "Дом", 100000, "Не назначен")

	// Пусто
	if p, total, c, err := GetProjectStageProgress(pid); err != nil || total != 0 || p != 0 || c != 0 {
		t.Fatalf("пустой прогресс этапов: p=%d total=%d c=%d err=%v", p, total, c, err)
	}

	// Добавляем этапы — позиции автоинкрементятся
	s1, err := AddStage(pid, "Фундамент")
	if err != nil {
		t.Fatalf("AddStage: %v", err)
	}
	AddStage(pid, "Стены")

	stages, _ := GetProjectStages(pid)
	if len(stages) != 2 {
		t.Fatalf("ожидалось 2 этапа, получили %d", len(stages))
	}
	if stages[0].Name != "Фундамент" || stages[0].Position != 1 {
		t.Errorf("первый этап неверен: %+v", stages[0])
	}
	if stages[1].Position != 2 {
		t.Errorf("позиция второго этапа = %d, want 2", stages[1].Position)
	}

	// Завершаем первый этап → прогресс 50%
	if err := UpdateStageStatus(s1, "completed"); err != nil {
		t.Fatalf("UpdateStageStatus: %v", err)
	}
	p, total, c, _ := GetProjectStageProgress(pid)
	if total != 2 || c != 1 || p != 50 {
		t.Errorf("прогресс этапов: p=%d total=%d c=%d, want 50/2/1", p, total, c)
	}

	// GetStageByID
	st, err := GetStageByID(s1)
	if err != nil || st.Status != "completed" {
		t.Errorf("GetStageByID: %+v, %v", st, err)
	}
}

func TestNextStageStatus(t *testing.T) {
	if NextStageStatus("pending") != "in_progress" {
		t.Error("pending → in_progress")
	}
	if NextStageStatus("in_progress") != "completed" {
		t.Error("in_progress → completed")
	}
	if NextStageStatus("completed") != "pending" {
		t.Error("completed → pending")
	}
}

func TestSeedStandardStages(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1201, "Строительство", "Дом", 1, "Не назначен")
	if err := SeedStandardStages(pid); err != nil {
		t.Fatalf("SeedStandardStages: %v", err)
	}
	stages, _ := GetProjectStages(pid)
	if len(stages) != len(StandardStages) {
		t.Errorf("ожидалось %d типовых этапов, получили %d", len(StandardStages), len(stages))
	}
	if stages[0].Name != "Фундамент" {
		t.Errorf("первый типовой этап = %q", stages[0].Name)
	}
}

func TestStagesValidationAndErrors(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1202, "Монтаж", "P", 1, "Не назначен")
	sid, _ := AddStage(pid, "Этап")

	if _, err := AddStage(999999, "нет проекта"); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта")
	}
	if err := UpdateStageStatus(sid, "странный"); err == nil {
		t.Error("ожидалась ошибка для неверного статуса")
	}
	if err := UpdateStageStatus(999999, "completed"); err == nil {
		t.Error("ожидалась ошибка для несуществующего этапа")
	}
	if _, err := GetStageByID(999999); err == nil {
		t.Error("ожидалась ошибка для несуществующего этапа")
	}
}

func TestStagesErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetProjectStages(1); err == nil {
		t.Error("GetProjectStages на закрытой БД должен вернуть ошибку")
	}
	if _, _, _, err := GetProjectStageProgress(1); err == nil {
		t.Error("GetProjectStageProgress на закрытой БД должен вернуть ошибку")
	}
}
