package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Роли пользователей.
const (
	RoleManager = "manager" // полный доступ
	RoleForeman = "foreman" // задачи, расходы, фото
	RoleViewer  = "viewer"  // только просмотр
)

// RoleName — человекочитаемое название роли.
func RoleName(role string) string {
	switch role {
	case RoleManager:
		return "Менеджер"
	case RoleForeman:
		return "Прораб"
	case RoleViewer:
		return "Заказчик (просмотр)"
	default:
		return "Менеджер"
	}
}

// IsValidRole — допустима ли роль.
func IsValidRole(role string) bool {
	return role == RoleManager || role == RoleForeman || role == RoleViewer
}

// RoleCanEdit — может ли роль работать с задачами/расходами/фото.
func RoleCanEdit(role string) bool {
	return role == RoleManager || role == RoleForeman
}

// RoleCanManage — может ли роль управлять проектами и мастерами.
func RoleCanManage(role string) bool {
	return role == RoleManager
}

// GetUserRole — роль пользователя (по умолчанию «менеджер»).
func GetUserRole(userID int64) (string, error) {
	if err := ensureUser(userID); err != nil {
		return RoleManager, err
	}

	var role sql.NullString
	err := DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return RoleManager, err
	}
	if !role.Valid || role.String == "" {
		return RoleManager, nil
	}
	return role.String, nil
}

// SetUserRole — устанавливает роль пользователя.
func SetUserRole(userID int64, role string) error {
	if !IsValidRole(role) {
		return fmt.Errorf("неизвестная роль: %s", role)
	}
	if err := ensureUser(userID); err != nil {
		return err
	}
	_, err := DB.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	if err != nil {
		log.Printf("❌ Ошибка установки роли: %v", err)
	}
	return err
}
