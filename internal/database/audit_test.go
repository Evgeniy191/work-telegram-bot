package database

import "testing"

func TestAuditLogWrappers(t *testing.T) {
	defer setupTestDB(t)()

	const actor = int64(700)
	pid, _ := CreateProject(actor, "Монтаж", "Объект", 10000, "Не назначен")
	tid, _ := CreateTask(pid, "Задача", "", nil)

	// Пустой журнал.
	if log, err := GetProjectAuditLog(pid, 20); err != nil || len(log) != 0 {
		t.Fatalf("пустой журнал: len=%d err=%v", len(log), err)
	}

	// Смена статуса → запись.
	if err := UpdateTaskStatusAudited(actor, "@boss", tid, "in_progress"); err != nil {
		t.Fatalf("UpdateTaskStatusAudited: %v", err)
	}
	// Повтор того же статуса → новой записи быть не должно.
	if err := UpdateTaskStatusAudited(actor, "@boss", tid, "in_progress"); err != nil {
		t.Fatalf("UpdateTaskStatusAudited (повтор): %v", err)
	}

	// Назначение исполнителя → запись.
	masters, _ := GetAllMasters()
	if len(masters) == 0 {
		t.Fatal("ожидались сидированные мастера")
	}
	mID := masters[0].ID
	if err := UpdateTaskAssigneeAudited(actor, "@boss", tid, &mID); err != nil {
		t.Fatalf("UpdateTaskAssigneeAudited: %v", err)
	}

	// Смена бюджета → запись.
	if err := UpdateProjectBudgetAudited(actor, "@boss", pid, 15000); err != nil {
		t.Fatalf("UpdateProjectBudgetAudited: %v", err)
	}
	// Тот же бюджет → без новой записи.
	if err := UpdateProjectBudgetAudited(actor, "@boss", pid, 15000); err != nil {
		t.Fatalf("UpdateProjectBudgetAudited (повтор): %v", err)
	}

	log, err := GetProjectAuditLog(pid, 20)
	if err != nil {
		t.Fatalf("GetProjectAuditLog: %v", err)
	}
	if len(log) != 3 {
		t.Fatalf("ожидалось 3 записи журнала, получили %d: %+v", len(log), log)
	}

	// Новые записи первыми: последней менялся бюджет.
	if log[0].Field != "budget" || log[0].OldValue != "10000.00" || log[0].NewValue != "15000.00" {
		t.Errorf("запись[0] (бюджет) неверна: %+v", log[0])
	}
	if log[1].Field != "assignee" || log[1].OldValue != "Не назначен" || log[1].NewValue != masters[0].Name {
		t.Errorf("запись[1] (исполнитель) неверна: %+v", log[1])
	}
	if log[2].Field != "status" || log[2].NewValue != "В процессе" {
		t.Errorf("запись[2] (статус) неверна: %+v", log[2])
	}
	if log[0].UserName != "@boss" || log[0].UserID != actor || log[0].ProjectID != pid {
		t.Errorf("метаданные автора неверны: %+v", log[0])
	}

	// Лимит ограничивает выборку.
	if limited, _ := GetProjectAuditLog(pid, 1); len(limited) != 1 {
		t.Errorf("лимит=1: ожидалась 1 запись, получили %d", len(limited))
	}
}

func TestAuditFieldName(t *testing.T) {
	cases := map[string]string{
		"status":   "статус",
		"budget":   "бюджет",
		"assignee": "исполнитель",
		"unknown":  "unknown",
	}
	for in, want := range cases {
		if got := AuditFieldName(in); got != want {
			t.Errorf("AuditFieldName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAuditWrappersMissingEntity(t *testing.T) {
	defer setupTestDB(t)()

	mID := int64(1)
	if err := UpdateTaskStatusAudited(1, "x", 999999, "completed"); err == nil {
		t.Error("ожидалась ошибка для несуществующей задачи (статус)")
	}
	if err := UpdateTaskAssigneeAudited(1, "x", 999999, &mID); err == nil {
		t.Error("ожидалась ошибка для несуществующей задачи (исполнитель)")
	}
	if err := UpdateProjectBudgetAudited(1, "x", 999999, 100); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта (бюджет)")
	}
}

func TestAuditErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if err := AddAuditEntry(1, "x", 1, "task", 1, "status", "a", "b"); err == nil {
		t.Error("AddAuditEntry на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetProjectAuditLog(1, 20); err == nil {
		t.Error("GetProjectAuditLog на закрытой БД должен вернуть ошибку")
	}
}
