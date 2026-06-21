package handlers

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestAuditCallbackView(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7900)
	pid, _ := database.CreateProject(chatID, "Монтаж", "Аудит-проект", 1000, "Не назначен")

	// Пустой журнал.
	cap := cb(t, chatID, fmt.Sprintf("audit_%d", pid), mgr)
	if !cap.contains("Журнал изменений") || !cap.contains("нет записей") {
		t.Errorf("ожидался пустой журнал, получили: %s", cap.joined())
	}

	// Невалидный ID и отсутствующий проект.
	if cap = cb(t, chatID, "audit_abc", mgr); !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	if cap = cb(t, chatID, "audit_999999", mgr); !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}

	// Смена статуса задачи через callback должна попасть в журнал.
	tid, _ := database.CreateTask(pid, "Задача", "", nil)
	cb(t, chatID, fmt.Sprintf("task_status_%d_completed", tid), mgr)

	cap = cb(t, chatID, fmt.Sprintf("audit_%d", pid), mgr)
	if !cap.contains("статус") || !cap.contains("Выполнена") {
		t.Errorf("ожидалась запись о смене статуса, получили: %s", cap.joined())
	}
}

func TestAuditLogText(t *testing.T) {
	p := &database.Project{ID: 1, Name: "Тест"}

	if got := auditLogText(p, nil); !strings.Contains(got, "Пока нет записей") {
		t.Errorf("пустой журнал: %s", got)
	}

	entries := []database.AuditEntry{
		{Field: "budget", OldValue: "100.00", NewValue: "200.00", UserName: "@boss", CreatedAt: time.Now()},
		{Field: "assignee", OldValue: "Не назначен", NewValue: "Иванов", UserID: 42, CreatedAt: time.Now()},
	}
	got := auditLogText(p, entries)
	for _, want := range []string{"бюджет", "100.00", "200.00", "@boss", "исполнитель", "Иванов", "ID 42"} {
		if !strings.Contains(got, want) {
			t.Errorf("в журнале нет %q\n%s", want, got)
		}
	}
}

func TestActorName(t *testing.T) {
	if got := actorName(nil); got != "" {
		t.Errorf("actorName(nil) = %q, want \"\"", got)
	}
	if got := actorName(&tgbotapi.User{UserName: "boss"}); got != "@boss" {
		t.Errorf("actorName(username) = %q, want @boss", got)
	}
	if got := actorName(&tgbotapi.User{FirstName: "Пётр"}); got != "Пётр" {
		t.Errorf("actorName(firstname) = %q, want Пётр", got)
	}
}
