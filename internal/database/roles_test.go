package database

import "testing"

func TestRolePermissionHelpers(t *testing.T) {
	if !RoleCanManage(RoleManager) {
		t.Error("менеджер должен иметь права управления")
	}
	if RoleCanManage(RoleForeman) || RoleCanManage(RoleViewer) {
		t.Error("прораб/заказчик не должны управлять проектами")
	}
	if !RoleCanEdit(RoleManager) || !RoleCanEdit(RoleForeman) {
		t.Error("менеджер и прораб должны редактировать")
	}
	if RoleCanEdit(RoleViewer) {
		t.Error("заказчик не должен редактировать")
	}

	if RoleName(RoleManager) != "Менеджер" || RoleName(RoleForeman) != "Прораб" {
		t.Error("неверные названия ролей")
	}
	if RoleName("странная") != "Менеджер" {
		t.Error("неизвестная роль должна давать дефолт")
	}

	if !IsValidRole(RoleViewer) || IsValidRole("xxx") {
		t.Error("IsValidRole работает неверно")
	}
}

func TestGetSetUserRole(t *testing.T) {
	defer setupTestDB(t)()

	uid := int64(1100)

	// По умолчанию — менеджер
	role, err := GetUserRole(uid)
	if err != nil {
		t.Fatalf("GetUserRole: %v", err)
	}
	if role != RoleManager {
		t.Errorf("роль по умолчанию = %q, want manager", role)
	}

	// Меняем на заказчика
	if err := SetUserRole(uid, RoleViewer); err != nil {
		t.Fatalf("SetUserRole: %v", err)
	}
	role, _ = GetUserRole(uid)
	if role != RoleViewer {
		t.Errorf("роль = %q, want viewer", role)
	}

	// Невалидная роль
	if err := SetUserRole(uid, "босс"); err == nil {
		t.Error("ожидалась ошибка для невалидной роли")
	}
}

func TestUserRoleErrorPath(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := GetUserRole(1); err == nil {
		t.Error("GetUserRole на закрытой БД должен вернуть ошибку")
	}
	if err := SetUserRole(1, RoleManager); err == nil {
		t.Error("SetUserRole на закрытой БД должен вернуть ошибку")
	}
}
