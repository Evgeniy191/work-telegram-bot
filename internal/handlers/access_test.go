package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestRequiredCallbackPermission(t *testing.T) {
	cases := map[string]permLevel{
		"project_new":          permManage,
		"delete_project_5":     permManage,
		"edit_project_5":       permManage,
		"edit_master_name_2":   permManage,
		"master_pick_3":        permManage,
		"add_task_5":           permEdit,
		"edit_task_5":          permEdit,
		"task_status_5_done":   permEdit,
		"add_expense_5":        permEdit,
		"add_photo_5":          permEdit,
		"task_assignee_5_2":    permEdit,
		"view_tasks_5":         permView,
		"expenses_5":           permView,
		"show_photos_5":        permView,
		"toggle_notifications": permView,
		"set_role_viewer":      permView,
	}
	for data, want := range cases {
		if got := requiredCallbackPermission(data); got != want {
			t.Errorf("requiredCallbackPermission(%q) = %d, want %d", data, got, want)
		}
	}
}

func TestAccessViewerBlocked(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7400)
	database.SetUserRole(chatID, database.RoleViewer)
	defer database.SetUserRole(chatID, database.RoleManager) // вернуть, чтобы не влиять на другие тесты

	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")
	tid, _ := database.CreateTask(pid, "T", "", nil)

	// Управление — запрещено
	cap := cb(t, chatID, fmt.Sprintf("delete_project_%d", pid), mgr)
	if !cap.contains("Недостаточно прав") {
		t.Errorf("заказчик не должен удалять проект, получили: %s", cap.joined())
	}

	// Редактирование — запрещено
	cap = cb(t, chatID, fmt.Sprintf("edit_task_%d", tid), mgr)
	if !cap.contains("Недостаточно прав") {
		t.Errorf("заказчик не должен редактировать задачу, получили: %s", cap.joined())
	}

	// Просмотр — разрешено
	cap = cb(t, chatID, fmt.Sprintf("view_tasks_%d", pid), mgr)
	if cap.contains("Недостаточно прав") {
		t.Errorf("заказчик должен иметь доступ к просмотру, получили: %s", cap.joined())
	}
}

func TestAccessForeman(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7401)
	database.SetUserRole(chatID, database.RoleForeman)
	defer database.SetUserRole(chatID, database.RoleManager)

	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 100, "")

	// Прораб может работать с задачами
	cap := cb(t, chatID, fmt.Sprintf("add_task_%d", pid), mgr)
	if cap.contains("Недостаточно прав") {
		t.Errorf("прораб должен добавлять задачи, получили: %s", cap.joined())
	}

	// Но не может удалять проект
	cap = cb(t, chatID, fmt.Sprintf("delete_project_%d", pid), mgr)
	if !cap.contains("Недостаточно прав") {
		t.Errorf("прораб не должен удалять проект, получили: %s", cap.joined())
	}
}

func TestSetRoleCallback(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7402)

	cap := cb(t, chatID, "set_role_viewer", mgr)
	if !cap.contains("Роль и доступ") {
		t.Errorf("ожидался экран роли, получили: %s", cap.joined())
	}
	role, _ := database.GetUserRole(chatID)
	if role != database.RoleViewer {
		t.Errorf("роль не изменилась: %q", role)
	}
	database.SetUserRole(chatID, database.RoleManager)

	cap = cb(t, chatID, "set_role_unknown", mgr)
	if !cap.contains("Неизвестная роль") {
		t.Errorf("ожидалась ошибка роли, получили: %s", cap.joined())
	}
}

func TestHandleSecurityRoleScreen(t *testing.T) {
	bot, cap := newTestBot(t)
	HandleSecurity(bot, msgUpdate(7403, "🔐 Безопасность"))
	if !cap.contains("Роль и доступ") {
		t.Errorf("ожидался экран роли, получили: %s", cap.joined())
	}
}

func TestRouteCommandViewerCannotCreate(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7404)
	database.SetUserRole(chatID, database.RoleViewer)
	defer database.SetUserRole(chatID, database.RoleManager)

	bot, cap := newTestBot(t)
	RouteCommand(bot, msgUpdate(chatID, "➕ Новый проект"), mgr)
	if !cap.contains("Недостаточно прав") {
		t.Errorf("заказчик не должен создавать проект, получили: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("состояние создания не должно устанавливаться для заказчика")
	}
}
