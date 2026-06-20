package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackExpensesMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7300)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 10000, "")
	database.AddExpense(pid, 2500, "материалы")

	cap := cb(t, chatID, fmt.Sprintf("expenses_%d", pid), mgr)
	if !cap.contains("план/факт") {
		t.Errorf("ожидалось меню бюджета, получили: %s", cap.joined())
	}
	if !cap.contains("материалы") {
		t.Errorf("ожидался список расходов, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "expenses_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "expenses_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackAddExpenseStart(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7301)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 10000, "")

	cb(t, chatID, fmt.Sprintf("add_expense_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingExpenseAmount {
		t.Error("add_expense не установил состояние")
	}
	if mgr.GetData(chatID).EditingProjectID != pid {
		t.Error("EditingProjectID не сохранён")
	}

	cap := cb(t, chatID, "add_expense_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "add_expense_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestFSMAddExpenseFlow(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7302)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 10000, "")

	mgr.SetState(chatID, fsm.StateAddingExpenseAmount)
	mgr.GetData(chatID).EditingProjectID = pid

	// Невалидная сумма
	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "много"), mgr)
	if !cap.contains("числом") {
		t.Errorf("ожидалась ошибка суммы, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateAddingExpenseAmount {
		t.Error("при ошибке состояние не должно меняться")
	}

	// Валидная сумма → переход к описанию
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "3000"), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingExpenseDesc {
		t.Errorf("ожидался переход к описанию, состояние=%s", mgr.GetState(chatID))
	}

	// Описание → сохранение
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "бетон"), mgr)
	if !cap.contains("Расход добавлен") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("после сохранения состояние должно сброситься")
	}

	spent, _ := database.GetProjectSpent(pid)
	if spent != 3000 {
		t.Errorf("расход не сохранён: spent=%v", spent)
	}
}

func TestFSMAddExpenseOverBudgetWarning(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7303)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1000, "")

	mgr.SetState(chatID, fsm.StateAddingExpenseAmount)
	mgr.GetData(chatID).EditingProjectID = pid

	bot, _ := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "5000"), mgr)

	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "/skip"), mgr)
	if !cap.contains("перерасход") {
		t.Errorf("ожидалось предупреждение о перерасходе, получили: %s", cap.joined())
	}
}

func TestMyProjectsShowsBudgetUsage(t *testing.T) {
	bot, cap := newTestBot(t)
	chatID := int64(7304)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 10000, "")
	database.AddExpense(pid, 5000, "")

	HandleMyProjects(bot, msgUpdate(chatID, "📋 Проекты"))
	if !cap.contains("освоено 50%") {
		t.Errorf("ожидался показатель освоения бюджета, получили: %s", cap.joined())
	}
}
