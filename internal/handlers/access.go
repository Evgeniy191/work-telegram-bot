package handlers

import (
	"fmt"
	"strings"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// permLevel — требуемый уровень прав для действия.
type permLevel int

const (
	permView   permLevel = iota // просмотр — доступен всем
	permEdit                    // задачи/расходы/фото — менеджер и прораб
	permManage                  // проекты/мастера — только менеджер
)

// requiredCallbackPermission классифицирует callback по требуемым правам.
func requiredCallbackPermission(data string) permLevel {
	switch {
	// Управление проектами и мастерами — только менеджер
	case data == "project_new",
		data == "save_project",
		data == "edit_last_project",
		data == "edit_type", data == "edit_name", data == "edit_budget", data == "edit_master",
		data == "master_none",
		strings.HasPrefix(data, "type_"),
		strings.HasPrefix(data, "master_pick_"),
		strings.HasPrefix(data, "edit_project_"),
		strings.HasPrefix(data, "edit_name_"),
		strings.HasPrefix(data, "edit_budget_"),
		strings.HasPrefix(data, "edit_master_"), // включая edit_master_name_
		strings.HasPrefix(data, "delete_project_"),
		strings.HasPrefix(data, "confirm_delete_"):
		return permManage

	// Работа с задачами/расходами/фото — менеджер и прораб
	case strings.HasPrefix(data, "add_task_"),
		strings.HasPrefix(data, "edit_task_"),
		strings.HasPrefix(data, "task_status_"),
		strings.HasPrefix(data, "delete_task_"),
		strings.HasPrefix(data, "task_assignee_"),
		strings.HasPrefix(data, "add_expense_"),
		strings.HasPrefix(data, "add_photo_"):
		return permEdit

	default:
		return permView
	}
}

// ensureAccess проверяет права пользователя; при недостатке прав отвечает отказом.
func ensureAccess(bot *tgbotapi.BotAPI, chatID int64, perm permLevel) bool {
	if perm == permView {
		return true
	}

	role, _ := database.GetUserRole(chatID)

	allowed := false
	switch perm {
	case permEdit:
		allowed = database.RoleCanEdit(role)
	case permManage:
		allowed = database.RoleCanManage(role)
	}

	if !allowed {
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf(
			"⛔ Недостаточно прав для этого действия.\nВаша роль: %s", database.RoleName(role))))
	}
	return allowed
}

// checkCallbackAccess — гейт доступа для callback-кнопок.
func checkCallbackAccess(bot *tgbotapi.BotAPI, chatID int64, data string) bool {
	return ensureAccess(bot, chatID, requiredCallbackPermission(data))
}

// roleScreenText — текст экрана роли и доступа.
func roleScreenText(role string) string {
	return fmt.Sprintf(
		"🔐 *Роль и доступ*\n\n"+
			"Текущая роль: *%s*\n\n"+
			"Уровни доступа:\n"+
			"• *Менеджер* — полный доступ (проекты, мастера)\n"+
			"• *Прораб* — задачи, расходы, фото\n"+
			"• *Заказчик* — только просмотр\n\n"+
			"Выбери роль:",
		database.RoleName(role))
}
