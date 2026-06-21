package database

import (
	"testing"
	"time"
)

func TestProjectCardFields(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1300, "Строительство", "Дом", 100000, "Не назначен")

	// По умолчанию поля пустые
	p, _ := GetProjectByID(pid)
	if p.Address != "" || p.Customer != "" || p.ContractNumber != "" || p.StartDate != nil || p.PlannedEnd != nil {
		t.Errorf("ожидались пустые поля карточки: %+v", p)
	}

	// Текстовые поля
	if err := UpdateProjectCardText(pid, "address", "ул. Строителей, 5"); err != nil {
		t.Fatalf("UpdateProjectCardText(address): %v", err)
	}
	UpdateProjectCardText(pid, "customer", "ООО Заказчик")
	UpdateProjectCardText(pid, "contract", "Д-2026/14")

	// Даты
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
	if err := UpdateProjectDate(pid, "start", &start); err != nil {
		t.Fatalf("UpdateProjectDate(start): %v", err)
	}
	UpdateProjectDate(pid, "end", &end)

	p, _ = GetProjectByID(pid)
	if p.Address != "ул. Строителей, 5" || p.Customer != "ООО Заказчик" || p.ContractNumber != "Д-2026/14" {
		t.Errorf("текстовые поля карточки неверны: %+v", p)
	}
	if p.StartDate == nil || p.StartDate.Format("02.01.2006") != "01.07.2026" {
		t.Errorf("дата старта неверна: %v", p.StartDate)
	}
	if p.PlannedEnd == nil || p.PlannedEnd.Format("02.01.2006") != "20.12.2026" {
		t.Errorf("плановая сдача неверна: %v", p.PlannedEnd)
	}

	// Поля приходят и в списке проектов
	list, _ := GetUserProjects(1300)
	if len(list) != 1 || list[0].Address != "ул. Строителей, 5" {
		t.Errorf("карточка не читается в GetUserProjects: %+v", list)
	}

	// Сброс даты
	if err := UpdateProjectDate(pid, "start", nil); err != nil {
		t.Fatalf("UpdateProjectDate(nil): %v", err)
	}
	p, _ = GetProjectByID(pid)
	if p.StartDate != nil {
		t.Error("дата старта должна сброситься")
	}
}

func TestProjectCardInvalidField(t *testing.T) {
	defer setupTestDB(t)()
	pid, _ := CreateProject(1301, "Монтаж", "P", 1, "Не назначен")

	if err := UpdateProjectCardText(pid, "нечто", "x"); err == nil {
		t.Error("ожидалась ошибка для неизвестного текстового поля")
	}
	if err := UpdateProjectDate(pid, "нечто", nil); err == nil {
		t.Error("ожидалась ошибка для неизвестного поля даты")
	}
}
