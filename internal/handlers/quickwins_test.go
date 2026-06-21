package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackEditTaskMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(6000)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "Описание", nil)

	cap := cb(t, chatID, fmt.Sprintf("edit_task_%d", tid), mgr)
	if !cap.contains("Редактирование задачи") {
		t.Errorf("ожидалось меню задачи, получили: %s", cap.joined())
	}

	// Неверный ID
	cap = cb(t, chatID, "edit_task_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}

	// Несуществующая задача
	cap = cb(t, chatID, "edit_task_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestCallbackEditTaskFields(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(6001)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "Описание", nil)

	// Название
	cb(t, chatID, fmt.Sprintf("edit_task_name_%d", tid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingTaskName {
		t.Error("edit_task_name_ не установил состояние")
	}
	if mgr.GetData(chatID).EditingTaskID != tid {
		t.Error("EditingTaskID не сохранён")
	}

	// Описание
	cb(t, chatID, fmt.Sprintf("edit_task_desc_%d", tid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingTaskDescription {
		t.Error("edit_task_desc_ не установил состояние")
	}

	// Срок
	cb(t, chatID, fmt.Sprintf("edit_task_deadline_%d", tid), mgr)
	if mgr.GetState(chatID) != fsm.StateEditingTaskDeadline {
		t.Error("edit_task_deadline_ не установил состояние")
	}

	// Ошибки разбора ID / отсутствие задачи
	cap := cb(t, chatID, "edit_task_name_xx", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "edit_task_name_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestFSMEditingTaskFields(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(6002)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Старое", "Старое описание", nil)

	// Название: пустое
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskName)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "  "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого имени, получили: %s", cap.joined())
	}

	// Название: корректное
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskName)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "Новое имя"), mgr)
	if !cap.contains("обновлено") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	task, _ := database.GetTaskByID(tid)
	if task.Name != "Новое имя" {
		t.Errorf("имя задачи не обновилось: %q", task.Name)
	}

	// Описание: /clear
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskDescription)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "/clear"), mgr)
	task, _ = database.GetTaskByID(tid)
	if task.Description != "" {
		t.Errorf("описание не очищено: %q", task.Description)
	}

	// Срок: неверный формат
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskDeadline)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "кривая дата"), mgr)
	if !cap.contains("формат") {
		t.Errorf("ожидалась ошибка формата, получили: %s", cap.joined())
	}

	// Срок: корректный
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskDeadline)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "01.06.2027"), mgr)
	if !cap.contains("Срок задачи обновлён") {
		t.Errorf("ожидалось подтверждение срока, получили: %s", cap.joined())
	}
	task, _ = database.GetTaskByID(tid)
	if task.Deadline == nil || task.Deadline.Format("02.01.2006") != "01.06.2027" {
		t.Errorf("срок не обновился: %v", task.Deadline)
	}

	// Срок: /clear
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateEditingTaskDeadline)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "/clear"), mgr)
	task, _ = database.GetTaskByID(tid)
	if task.Deadline != nil {
		t.Error("срок должен быть убран через /clear")
	}
}

func TestHandleReports(t *testing.T) {
	// Пустой портфель
	bot, cap := newTestBot(t)
	HandleReports(bot, msgUpdate(6100, "📊 Отчёты"))
	if !cap.contains("нет проектов") {
		t.Errorf("ожидалось сообщение об отсутствии проектов, получили: %s", cap.joined())
	}

	// С данными
	chatID := int64(6101)
	pid, _ := dbCreateProjectWithTasks(chatID, 4, 2)
	past := past72h()
	tasks, _ := database.GetProjectTasks(pid)
	database.UpdateTaskDeadline(tasks[3].ID, &past) // одна просрочка

	database.AddExpense(pid, 250, "материалы")

	bot, cap = newTestBot(t)
	HandleReports(bot, msgUpdate(chatID, "📊 Отчёты"))
	if !cap.contains("Дашборд портфеля") {
		t.Errorf("ожидался дашборд, получили: %s", cap.joined())
	}
	if !cap.contains("50%") {
		t.Errorf("ожидался прогресс 50%%, получили: %s", cap.joined())
	}
	if !cap.contains("В работе") {
		t.Errorf("ожидалась разбивка по статусам, получили: %s", cap.joined())
	}
	if !cap.contains("Освоено") {
		t.Errorf("ожидалось освоение бюджета, получили: %s", cap.joined())
	}
}

func TestNotificationsToggle(t *testing.T) {
	chatID := int64(6200)

	// Экран уведомлений
	bot, cap := newTestBot(t)
	HandleNotifications(bot, msgUpdate(chatID, "🔔 Уведомления"))
	if !cap.contains("Уведомления") {
		t.Errorf("ожидался экран уведомлений, получили: %s", cap.joined())
	}

	// По умолчанию включены
	s, _ := database.GetUserSettings(chatID)
	if !s.Notifications {
		t.Fatal("по умолчанию уведомления включены")
	}

	// Переключаем через callback
	mgr := fsm.NewManager()
	cap = cb(t, chatID, "toggle_notifications", mgr)
	if cap.sends() == 0 {
		t.Error("toggle_notifications не ответил")
	}
	s, _ = database.GetUserSettings(chatID)
	if s.Notifications {
		t.Error("уведомления должны были выключиться")
	}
}
