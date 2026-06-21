package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestAcceptanceFlowAccept(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7700)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)

	// Прораб сдаёт на проверку
	cb(t, chatID, fmt.Sprintf("task_status_%d_review", tid), mgr)
	task, _ := database.GetTaskByID(tid)
	if task.Status != "review" {
		t.Fatalf("статус после сдачи = %q, want review", task.Status)
	}

	// Менеджер принимает
	cap := cb(t, chatID, fmt.Sprintf("task_accept_%d", tid), mgr)
	if !cap.contains("принята") {
		t.Errorf("ожидалось 'принята', получили: %s", cap.joined())
	}
	task, _ = database.GetTaskByID(tid)
	if task.Status != "completed" {
		t.Errorf("после приёмки статус = %q, want completed", task.Status)
	}
}

func TestAcceptanceFlowReject(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7701)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)
	database.UpdateTaskStatus(tid, "review")

	cap := cb(t, chatID, fmt.Sprintf("task_reject_%d", tid), mgr)
	if !cap.contains("на доработку") {
		t.Errorf("ожидалось 'на доработку', получили: %s", cap.joined())
	}
	task, _ := database.GetTaskByID(tid)
	if task.Status != "in_progress" {
		t.Errorf("после возврата статус = %q, want in_progress", task.Status)
	}
}

func TestAcceptanceGuards(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7702)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	tid, _ := database.CreateTask(pid, "T", "", nil) // статус pending, не review

	// Принять задачу не на проверке — нельзя
	cap := cb(t, chatID, fmt.Sprintf("task_accept_%d", tid), mgr)
	if !cap.contains("не на проверке") {
		t.Errorf("ожидался отказ приёмки, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, fmt.Sprintf("task_reject_%d", tid), mgr)
	if !cap.contains("не на проверке") {
		t.Errorf("ожидался отказ возврата, получили: %s", cap.joined())
	}

	// Ошибки ID / отсутствие задачи
	cap = cb(t, chatID, "task_accept_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "task_accept_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "task_reject_xyz", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "task_reject_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestViewTasksReviewButtons(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7703)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	tid, _ := database.CreateTask(pid, "На проверку", "", nil)
	database.UpdateTaskStatus(tid, "review")

	cap := cb(t, chatID, fmt.Sprintf("view_tasks_%d", pid), mgr)
	if !cap.contains("На проверке") {
		t.Errorf("ожидался статус 'На проверке' в списке, получили: %s", cap.joined())
	}
}

func TestAcceptanceForemanCannotAccept(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7704)
	database.SetUserRole(chatID, database.RoleForeman)
	defer database.SetUserRole(chatID, database.RoleManager)

	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)
	database.UpdateTaskStatus(tid, "review")

	// Прораб не может принимать работу (только менеджер)
	cap := cb(t, chatID, fmt.Sprintf("task_accept_%d", tid), mgr)
	if !cap.contains("Недостаточно прав") {
		t.Errorf("прораб не должен принимать работу, получили: %s", cap.joined())
	}
}
