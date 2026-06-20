package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestViewTasksOverdueMarker(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7000)

	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "Просроченная", "", nil)
	past := past72h()
	database.UpdateTaskDeadline(tid, &past)

	cap := cb(t, chatID, fmt.Sprintf("view_tasks_%d", pid), mgr)
	if !cap.contains("ПРОСРОЧЕНО") {
		t.Errorf("ожидалась отметка просрочки, получили: %s", cap.joined())
	}
}

func TestMyProjectsOverdueCount(t *testing.T) {
	bot, cap := newTestBot(t)
	chatID := int64(7001)

	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)
	past := past72h()
	database.UpdateTaskDeadline(tid, &past)

	HandleMyProjects(bot, msgUpdate(chatID, "📋 Проекты"))
	if !cap.contains("Просрочено задач") {
		t.Errorf("ожидался счётчик просрочек в карточке, получили: %s", cap.joined())
	}
}
