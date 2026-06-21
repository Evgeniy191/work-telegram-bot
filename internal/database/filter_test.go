package database

import (
	"testing"
	"time"
)

func masterIDByName(t *testing.T, name string) int64 {
	t.Helper()
	masters, _ := GetAllMasters()
	for _, m := range masters {
		if m.Name == name {
			return m.ID
		}
	}
	t.Fatalf("мастер %q не найден среди сидированных", name)
	return 0
}

func TestFilterProjects(t *testing.T) {
	defer setupTestDB(t)()

	uid := int64(800)

	// p1: мастер Иванов Иван, объект «Москва», есть просроченная задача.
	p1, _ := CreateProject(uid, "Монтаж", "Альфа", 3000, "Иванов Иван")
	UpdateProjectCardText(p1, "address", "Москва, Ленина 1")
	to, _ := CreateTask(p1, "просрочка", "", nil)
	past := time.Now().Add(-24 * time.Hour)
	UpdateTaskDeadline(to, &past)

	// p2: без мастера, заказчик «ООО Тверь», без просрочки.
	p2, _ := CreateProject(uid, "Ремонт", "Бета", 1000, "Не назначен")
	UpdateProjectCardText(p2, "customer", "ООО Тверь")

	// Чужой проект — не должен попадать в выборку.
	CreateProject(999, "Монтаж", "Чужой", 1, "Не назначен")

	// Без фильтра — оба проекта пользователя.
	if all, _ := FilterProjects(uid, ProjectFilter{}); len(all) != 2 {
		t.Fatalf("без фильтра: %d, want 2", len(all))
	}

	// Только с просрочкой → p1.
	if ov, _ := FilterProjects(uid, ProjectFilter{OnlyOverdue: true}); len(ov) != 1 || ov[0].ID != p1 {
		t.Errorf("просрочка: %+v", ov)
	}

	// По мастеру Иванов → p1.
	ivan := masterIDByName(t, "Иванов Иван")
	if bm, _ := FilterProjects(uid, ProjectFilter{MasterID: &ivan}); len(bm) != 1 || bm[0].ID != p1 {
		t.Errorf("по мастеру: %+v", bm)
	}

	// Поиск по заказчику → p2.
	if s, _ := FilterProjects(uid, ProjectFilter{Search: "Тверь"}); len(s) != 1 || s[0].ID != p2 {
		t.Errorf("поиск по заказчику: %+v", s)
	}
	// Поиск по названию → p1.
	if s, _ := FilterProjects(uid, ProjectFilter{Search: "Альфа"}); len(s) != 1 || s[0].ID != p1 {
		t.Errorf("поиск по названию: %+v", s)
	}
	// Поиск без совпадений → пусто.
	if s, _ := FilterProjects(uid, ProjectFilter{Search: "несуществующее"}); len(s) != 0 {
		t.Errorf("поиск без совпадений: %+v", s)
	}

	// Сортировка по названию (А→Б): Альфа, Бета.
	if bn, _ := FilterProjects(uid, ProjectFilter{Sort: "name"}); len(bn) != 2 || bn[0].Name != "Альфа" || bn[1].Name != "Бета" {
		t.Errorf("сортировка по названию: %+v", bn)
	}
	// Сортировка по бюджету (убыв.): 3000, 1000.
	if bb, _ := FilterProjects(uid, ProjectFilter{Sort: "budget"}); len(bb) != 2 || bb[0].Budget != 3000 || bb[1].Budget != 1000 {
		t.Errorf("сортировка по бюджету: %+v", bb)
	}
}

func TestFilterMenuData(t *testing.T) {
	defer setupTestDB(t)()

	uid := int64(801)
	CreateProject(uid, "Монтаж", "A", 1, "Иванов Иван")
	CreateProject(uid, "Ремонт", "B", 1, "Не назначен")

	statuses, err := GetProjectStatuses(uid)
	if err != nil {
		t.Fatalf("GetProjectStatuses: %v", err)
	}
	if len(statuses) != 1 || statuses[0] != "В работе" {
		t.Errorf("статусы: %+v, want [В работе]", statuses)
	}

	masters, err := GetProjectMasters(uid)
	if err != nil {
		t.Fatalf("GetProjectMasters: %v", err)
	}
	if len(masters) != 1 || masters[0].Name != "Иванов Иван" {
		t.Errorf("мастера проектов: %+v, want [Иванов Иван]", masters)
	}
}

func TestFilterErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := FilterProjects(1, ProjectFilter{OnlyOverdue: true}); err == nil {
		t.Error("FilterProjects на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetProjectStatuses(1); err == nil {
		t.Error("GetProjectStatuses на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetProjectMasters(1); err == nil {
		t.Error("GetProjectMasters на закрытой БД должен вернуть ошибку")
	}
}
