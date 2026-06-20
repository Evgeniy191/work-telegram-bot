package fsm

import "testing"

func TestGetStateDefaultsToIdle(t *testing.T) {
	m := NewManager()
	if got := m.GetState(1); got != StateIdle {
		t.Errorf("GetState для нового пользователя = %q, want %q", got, StateIdle)
	}
}

func TestSetAndGetState(t *testing.T) {
	m := NewManager()
	m.SetState(7, StateCreatingProject)
	if got := m.GetState(7); got != StateCreatingProject {
		t.Errorf("GetState = %q, want %q", got, StateCreatingProject)
	}
}

func TestGetDataCreatesEmpty(t *testing.T) {
	m := NewManager()
	d := m.GetData(5)
	if d == nil {
		t.Fatal("GetData вернул nil")
	}
	if d.ProjectName != "" || d.EditingMasterID != 0 {
		t.Errorf("ожидались пустые данные, получили %+v", d)
	}

	// Повторный вызов возвращает тот же объект
	d.ProjectName = "Проект"
	d2 := m.GetData(5)
	if d2.ProjectName != "Проект" {
		t.Errorf("данные не сохранились между вызовами: %q", d2.ProjectName)
	}
}

func TestSetData(t *testing.T) {
	m := NewManager()
	m.SetData(9, &UserData{ProjectName: "X", EditingMasterID: 3})

	d := m.GetData(9)
	if d.ProjectName != "X" || d.EditingMasterID != 3 {
		t.Errorf("SetData не сработал: %+v", d)
	}
}

func TestResetState(t *testing.T) {
	m := NewManager()
	m.SetState(2, StateEditingMasterName)
	m.SetData(2, &UserData{ProjectName: "ToReset"})

	m.ResetState(2)

	if got := m.GetState(2); got != StateIdle {
		t.Errorf("после ResetState состояние = %q, want %q", got, StateIdle)
	}
	if d := m.GetData(2); d.ProjectName != "" {
		t.Errorf("после ResetState данные не очищены: %+v", d)
	}
}

func TestIsolationBetweenUsers(t *testing.T) {
	m := NewManager()
	m.SetState(1, StateCreatingTask)
	m.SetState(2, StateEditingProjectBudget)

	if m.GetState(1) != StateCreatingTask {
		t.Error("состояние пользователя 1 повреждено")
	}
	if m.GetState(2) != StateEditingProjectBudget {
		t.Error("состояние пользователя 2 повреждено")
	}
}
