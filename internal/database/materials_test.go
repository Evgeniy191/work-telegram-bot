package database

import "testing"

func TestMaterialsCRUD(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1500, "Строительство", "Дом", 1, "Не назначен")

	// Пусто
	if m, err := GetProjectMaterials(pid); err != nil || len(m) != 0 {
		t.Fatalf("GetProjectMaterials(пусто): %d, %v", len(m), err)
	}

	// Добавляем
	mid, err := AddMaterial(pid, "Цемент М500", "20 мешков")
	if err != nil {
		t.Fatalf("AddMaterial: %v", err)
	}
	AddMaterial(pid, "Арматура", "")

	materials, _ := GetProjectMaterials(pid)
	if len(materials) != 2 {
		t.Fatalf("ожидалось 2 материала, получили %d", len(materials))
	}
	if materials[0].Name != "Цемент М500" || materials[0].Quantity != "20 мешков" || materials[0].Status != "ordered" {
		t.Errorf("первый материал неверен: %+v", materials[0])
	}

	// Статус по циклу
	if err := UpdateMaterialStatus(mid, "in_transit"); err != nil {
		t.Fatalf("UpdateMaterialStatus: %v", err)
	}
	m, _ := GetMaterialByID(mid)
	if m.Status != "in_transit" {
		t.Errorf("статус = %q, want in_transit", m.Status)
	}
}

func TestMaterialStatusHelpers(t *testing.T) {
	if NextMaterialStatus("ordered") != "in_transit" ||
		NextMaterialStatus("in_transit") != "received" ||
		NextMaterialStatus("received") != "ordered" {
		t.Error("цикл статусов материала неверен")
	}
	if MaterialStatusEmoji("ordered") != "🛒" || MaterialStatusEmoji("received") != "✅" || MaterialStatusEmoji("x") != "❓" {
		t.Error("эмодзи статусов материала неверны")
	}
	if MaterialStatusName("in_transit") != "В пути" || MaterialStatusName("x") != "Неизвестно" {
		t.Error("названия статусов материала неверны")
	}
}

func TestMaterialsValidationAndErrors(t *testing.T) {
	defer setupTestDB(t)()
	pid, _ := CreateProject(1501, "Монтаж", "P", 1, "Не назначен")
	mid, _ := AddMaterial(pid, "Гвозди", "1 кг")

	if _, err := AddMaterial(999999, "нет проекта", ""); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта")
	}
	if err := UpdateMaterialStatus(mid, "странный"); err == nil {
		t.Error("ожидалась ошибка для неверного статуса")
	}
	if err := UpdateMaterialStatus(999999, "received"); err == nil {
		t.Error("ожидалась ошибка для несуществующего материала")
	}
	if _, err := GetMaterialByID(999999); err == nil {
		t.Error("ожидалась ошибка для несуществующего материала")
	}
}

func TestMaterialsErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetProjectMaterials(1); err == nil {
		t.Error("GetProjectMaterials на закрытой БД должен вернуть ошибку")
	}
	if _, err := AddMaterial(1, "x", ""); err == nil {
		t.Error("AddMaterial на закрытой БД должен вернуть ошибку")
	}
}
