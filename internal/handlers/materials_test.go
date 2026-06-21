package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackMaterialsMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7900)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	cap := cb(t, chatID, fmt.Sprintf("materials_%d", pid), mgr)
	if !cap.contains("Материалы проекта") {
		t.Errorf("ожидалось меню материалов, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "materials_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "materials_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackAddMaterialFlow(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7901)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")

	// Старт
	cb(t, chatID, fmt.Sprintf("add_material_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingMaterialName {
		t.Error("add_material не установил состояние")
	}
	if mgr.GetData(chatID).EditingProjectID != pid {
		t.Error("EditingProjectID не сохранён")
	}

	// Пустое имя
	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "  "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого имени, получили: %s", cap.joined())
	}

	// Имя → переход к количеству
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "Кирпич"), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingMaterialQty {
		t.Errorf("ожидался переход к количеству, состояние=%s", mgr.GetState(chatID))
	}

	// Количество → заявка создана
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "1000 шт"), mgr)
	if !cap.contains("Заявка добавлена") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	materials, _ := database.GetProjectMaterials(pid)
	if len(materials) != 1 || materials[0].Name != "Кирпич" || materials[0].Quantity != "1000 шт" {
		t.Errorf("материал не сохранён: %+v", materials)
	}

	// Ошибка ID
	cap = cb(t, chatID, "add_material_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
}

func TestCallbackMaterialStatusCycle(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7902)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	mid, _ := database.AddMaterial(pid, "Песок", "5 т")

	cap := cb(t, chatID, fmt.Sprintf("material_next_%d", mid), mgr)
	if !cap.contains("Материалы проекта") {
		t.Errorf("ожидалось обновлённое меню, получили: %s", cap.joined())
	}
	m, _ := database.GetMaterialByID(mid)
	if m.Status != "in_transit" {
		t.Errorf("статус после цикла = %q, want in_transit", m.Status)
	}

	cap = cb(t, chatID, "material_next_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "material_next_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestMaterialSkipQuantity(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7903)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")

	mgr.SetState(chatID, fsm.StateAddingMaterialName)
	mgr.GetData(chatID).EditingProjectID = pid

	bot, _ := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "Краска"), mgr)

	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "/skip"), mgr)
	if !cap.contains("Заявка добавлена") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	materials, _ := database.GetProjectMaterials(pid)
	if len(materials) != 1 || materials[0].Quantity != "" {
		t.Errorf("материал без количества сохранён неверно: %+v", materials)
	}
}
