package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackTaskPhotosMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7200)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)

	cap := cb(t, chatID, fmt.Sprintf("task_photos_%d", tid), mgr)
	if !cap.contains("Фотофиксация") {
		t.Errorf("ожидалось меню фото, получили: %s", cap.joined())
	}

	cap = cb(t, chatID, "task_photos_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "task_photos_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestCallbackAddPhoto(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7201)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)

	cb(t, chatID, fmt.Sprintf("add_photo_%d", tid), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingTaskPhoto {
		t.Error("add_photo не установил состояние ожидания фото")
	}
	if mgr.GetData(chatID).EditingTaskID != tid {
		t.Error("EditingTaskID не сохранён")
	}

	cap := cb(t, chatID, "add_photo_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "add_photo_999999", mgr)
	if !cap.contains("не найдена") {
		t.Errorf("ожидалась ошибка 'не найдена', получили: %s", cap.joined())
	}
}

func TestFSMAddTaskPhoto(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7202)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)

	// Не фото (текст) → ошибка, остаёмся в состоянии
	bot, cap := newTestBot(t)
	mgr.SetState(chatID, fsm.StateAddingTaskPhoto)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, msgUpdate(chatID, "просто текст"), mgr)
	if !cap.contains("не фото") {
		t.Errorf("ожидалась ошибка 'не фото', получили: %s", cap.joined())
	}

	// Фото → сохраняется
	bot, cap = newTestBot(t)
	mgr.SetState(chatID, fsm.StateAddingTaskPhoto)
	mgr.GetData(chatID).EditingTaskID = tid
	HandleFSMMessage(bot, photoUpdate(chatID, "file_big", "акт приёмки"), mgr)
	if !cap.contains("Фото добавлено") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("после добавления фото состояние должно сброситься")
	}

	photos, _ := database.GetTaskPhotos(tid)
	if len(photos) != 1 || photos[0].FileID != "file_big" || photos[0].Caption != "акт приёмки" {
		t.Errorf("фото сохранено неверно: %+v", photos)
	}
}

func TestCallbackShowPhotos(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7203)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Задача", "", nil)

	// Пусто
	cap := cb(t, chatID, fmt.Sprintf("show_photos_%d", tid), mgr)
	if !cap.contains("Фото пока нет") {
		t.Errorf("ожидалось 'Фото пока нет', получили: %s", cap.joined())
	}

	// С фото → отправляется sendPhoto
	database.AddTaskPhoto(tid, "file_a", "")
	database.AddTaskPhoto(tid, "file_b", "подпись")
	cap = cb(t, chatID, fmt.Sprintf("show_photos_%d", tid), mgr)
	if cap.method("sendPhoto") != 2 {
		t.Errorf("ожидалось 2 sendPhoto, получили %d", cap.method("sendPhoto"))
	}

	// Неверный ID
	cap = cb(t, chatID, "show_photos_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
}
