package keyboards

import (
	"os"
	"strings"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
)

// TestMain поднимает временную БД, т.к. клавиатуры мастеров читают из неё.
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "kb-test-*")
	if err != nil {
		panic(err)
	}
	cwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		panic(err)
	}
	// После открытия соединения возвращаем cwd, чтобы coverage-профиль
	// писался в исходную директорию пакета.
	if err := database.InitDB(); err != nil {
		panic(err)
	}
	os.Chdir(cwd)

	code := m.Run()

	database.CloseDB()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func TestReplyKeyboards(t *testing.T) {
	if len(MainMenu().Keyboard) == 0 {
		t.Error("MainMenu пустая")
	}
	if len(BackToMainMenu().Keyboard) == 0 {
		t.Error("BackToMainMenu пустая")
	}
	if len(QuickActionsMenu().Keyboard) == 0 {
		t.Error("QuickActionsMenu пустая")
	}
	if len(SettingsMenu().Keyboard) == 0 {
		t.Error("SettingsMenu пустая")
	}
	if !RemoveKeyboard().RemoveKeyboard {
		t.Error("RemoveKeyboard должен удалять клавиатуру")
	}
}

func TestStaticInlineKeyboards(t *testing.T) {
	if len(ProjectsList().InlineKeyboard) == 0 {
		t.Error("ProjectsList пустая")
	}
	if len(TasksList().InlineKeyboard) == 0 {
		t.Error("TasksList пустая")
	}
	if len(ConfirmAction().InlineKeyboard) == 0 {
		t.Error("ConfirmAction пустая")
	}
	if len(ProjectTypeKeyboard().InlineKeyboard) == 0 {
		t.Error("ProjectTypeKeyboard пустая")
	}

	back := BackButton("test_cb")
	if len(back.InlineKeyboard) == 0 || len(back.InlineKeyboard[0]) == 0 {
		t.Fatal("BackButton пустая")
	}
	if got := *back.InlineKeyboard[0][0].CallbackData; got != "test_cb" {
		t.Errorf("callback BackButton = %q, want test_cb", got)
	}
}

func TestMastersKeyboard(t *testing.T) {
	kb := MastersKeyboard()
	// 4 сидированных мастера (2 ряда) + ряд "Без мастера" + ряд "Назад"
	if len(kb.InlineKeyboard) < 3 {
		t.Errorf("MastersKeyboard: ожидалось >=3 рядов, получили %d", len(kb.InlineKeyboard))
	}

	// Последний ряд — кнопка "Назад" с callback back_to_budget
	last := kb.InlineKeyboard[len(kb.InlineKeyboard)-1]
	if got := *last[0].CallbackData; got != "back_to_budget" {
		t.Errorf("последняя кнопка = %q, want back_to_budget", got)
	}

	// Где-то должны быть динамические master_pick_*
	foundPick := false
	for _, row := range kb.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData != nil && strings.HasPrefix(*b.CallbackData, "master_pick_") {
				foundPick = true
			}
		}
	}
	if !foundPick {
		t.Error("не найдено ни одной кнопки master_pick_<id>")
	}
}

func TestTaskAssigneeKeyboard(t *testing.T) {
	kb := TaskAssigneeKeyboard(42)
	if len(kb.InlineKeyboard) < 2 {
		t.Fatalf("ожидалось >=2 рядов, получили %d", len(kb.InlineKeyboard))
	}

	foundPick := false
	foundUnassign := false
	for _, row := range kb.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData == nil {
				continue
			}
			switch {
			case strings.HasPrefix(*b.CallbackData, "task_assignee_42_") && !strings.HasSuffix(*b.CallbackData, "_0"):
				foundPick = true
			case *b.CallbackData == "task_assignee_42_0":
				foundUnassign = true
			}
		}
	}
	if !foundPick {
		t.Error("не найдено кнопок назначения мастера")
	}
	if !foundUnassign {
		t.Error("не найдено кнопки снятия исполнителя")
	}
}

func TestNotificationsKeyboard(t *testing.T) {
	on := NotificationsKeyboard(true)
	if len(on.InlineKeyboard) == 0 || len(on.InlineKeyboard[0]) == 0 {
		t.Fatal("NotificationsKeyboard(true) пустая")
	}
	if *on.InlineKeyboard[0][0].CallbackData != "toggle_notifications" {
		t.Error("ожидался callback toggle_notifications")
	}
	if !strings.Contains(on.InlineKeyboard[0][0].Text, "Выключить") {
		t.Errorf("для включённых ожидалось 'Выключить', получили %q", on.InlineKeyboard[0][0].Text)
	}

	off := NotificationsKeyboard(false)
	if !strings.Contains(off.InlineKeyboard[0][0].Text, "Включить") {
		t.Errorf("для выключенных ожидалось 'Включить', получили %q", off.InlineKeyboard[0][0].Text)
	}
}

func TestMastersManageKeyboard(t *testing.T) {
	kb := MastersManageKeyboard()
	if len(kb.InlineKeyboard) < 2 {
		t.Errorf("MastersManageKeyboard: ожидалось >=2 рядов, получили %d", len(kb.InlineKeyboard))
	}

	foundEdit := false
	for _, row := range kb.InlineKeyboard {
		for _, b := range row {
			if b.CallbackData != nil && strings.HasPrefix(*b.CallbackData, "edit_master_name_") {
				foundEdit = true
			}
		}
	}
	if !foundEdit {
		t.Error("не найдено ни одной кнопки edit_master_name_<id>")
	}
}
