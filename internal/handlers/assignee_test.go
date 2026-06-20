package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackEditTaskAssigneePicker(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7100)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)

	cap := cb(t, chatID, fmt.Sprintf("edit_task_assignee_%d", tid), mgr)
	if !cap.contains("исполнителя") {
		t.Errorf("ожидался выбор исполнителя, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "edit_task_assignee_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "edit_task_assignee_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestCallbackTaskAssigneeSet(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7101)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)
	masters, _ := database.GetAllMasters()
	mID := masters[0].ID

	// Назначение
	cap := cb(t, chatID, fmt.Sprintf("task_assignee_%d_%d", tid, mID), mgr)
	if !cap.contains("Исполнитель задачи") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	task, _ := database.GetTaskByID(tid)
	if task.AssigneeID == nil || *task.AssigneeID != mID {
		t.Errorf("исполнитель не сохранён: %v", task.AssigneeID)
	}

	// Снятие (masterID = 0)
	cap = cb(t, chatID, fmt.Sprintf("task_assignee_%d_0", tid), mgr)
	if !cap.contains("Не назначен") {
		t.Errorf("ожидалось снятие исполнителя, получили: %s", cap.joined())
	}
	task, _ = database.GetTaskByID(tid)
	if task.AssigneeID != nil {
		t.Error("исполнитель должен быть снят")
	}
}

func TestCallbackTaskAssigneeErrors(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7102)

	// Нет второй части
	cap := cb(t, chatID, "task_assignee_5", mgr)
	if !cap.contains("формат") {
		t.Errorf("ожидалась ошибка формата, получили: %s", cap.joined())
	}

	// Нечисловые ID
	cap = cb(t, chatID, "task_assignee_a_b", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующий мастер
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)
	cap = cb(t, chatID, fmt.Sprintf("task_assignee_%d_999999", tid), mgr)
	if !cap.contains("Мастер не найден") {
		t.Errorf("ожидалась ошибка 'Мастер не найден', получили: %s", cap.joined())
	}
}

func TestEditTaskMenuShowsAssignee(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7103)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "Описание", nil)

	cap := cb(t, chatID, fmt.Sprintf("edit_task_%d", tid), mgr)
	if !cap.contains("Исполнитель") {
		t.Errorf("меню задачи должно показывать исполнителя, получили: %s", cap.joined())
	}
}
