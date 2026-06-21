package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackCardMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7600)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	cap := cb(t, chatID, fmt.Sprintf("cardmenu_%d", pid), mgr)
	if !cap.contains("Карточка объекта") {
		t.Errorf("ожидалась карточка, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "cardmenu_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "cardmenu_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackCardSetStartsEdit(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7601)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	cb(t, chatID, fmt.Sprintf("cardset_address_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingProjectCard {
		t.Error("cardset не установил состояние")
	}
	if mgr.GetData(chatID).CardField != "address" || mgr.GetData(chatID).EditingProjectID != pid {
		t.Errorf("данные правки карточки не сохранены: %+v", mgr.GetData(chatID))
	}

	// Ошибки
	cap := cb(t, chatID, "cardset_address", mgr)
	if !cap.contains("Неверный формат") {
		t.Errorf("ожидалась ошибка формата, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "cardset_address_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "cardset_address_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestFSMEditCardText(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7602)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	mgr.SetState(chatID, fsm.StateEditingProjectCard)
	d := mgr.GetData(chatID)
	d.EditingProjectID = pid
	d.CardField = "customer"

	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "ООО Ромашка"), mgr)
	if !cap.contains("Карточка объекта обновлена") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	p, _ := database.GetProjectByID(pid)
	if p.Customer != "ООО Ромашка" {
		t.Errorf("заказчик не сохранён: %q", p.Customer)
	}
}

func TestFSMEditCardDate(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7603)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	// Неверный формат даты
	mgr.SetState(chatID, fsm.StateEditingProjectCard)
	d := mgr.GetData(chatID)
	d.EditingProjectID = pid
	d.CardField = "end"

	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "не дата"), mgr)
	if !cap.contains("формат") {
		t.Errorf("ожидалась ошибка формата, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateEditingProjectCard {
		t.Error("при ошибке состояние не должно меняться")
	}

	// Корректная дата
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "20.12.2026"), mgr)
	if !cap.contains("обновлена") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	p, _ := database.GetProjectByID(pid)
	if p.PlannedEnd == nil || p.PlannedEnd.Format("02.01.2006") != "20.12.2026" {
		t.Errorf("плановая сдача не сохранена: %v", p.PlannedEnd)
	}
}
