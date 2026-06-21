package database

import (
	"fmt"
	"log"
	"time"
)

// AuditEntry — запись журнала изменений: кто, когда и что изменил.
type AuditEntry struct {
	ID        int64
	UserID    int64  // Telegram ID автора изменения
	UserName  string // отображаемое имя автора на момент изменения
	ProjectID int64  // проект, к которому относится изменение
	Entity    string // 'project' | 'task'
	EntityID  int64  // ID проекта или задачи
	Field     string // 'status' | 'budget' | 'assignee'
	OldValue  string
	NewValue  string
	CreatedAt time.Time
}

// AddAuditEntry — добавляет запись в журнал изменений.
func AddAuditEntry(userID int64, userName string, projectID int64, entity string, entityID int64, field, oldVal, newVal string) error {
	_, err := DB.Exec(`
		INSERT INTO audit_log (user_id, user_name, project_id, entity, entity_id, field, old_value, new_value, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
	`, userID, userName, projectID, entity, entityID, field, oldVal, newVal)
	if err != nil {
		log.Printf("❌ Ошибка записи в журнал изменений: %v", err)
	}
	return err
}

// GetProjectAuditLog — последние записи журнала по проекту (новые первыми).
// limit <= 0 трактуется как значение по умолчанию (20).
func GetProjectAuditLog(projectID int64, limit int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := DB.Query(`
		SELECT id, user_id, IFNULL(user_name, ''), project_id, entity, entity_id,
		       field, IFNULL(old_value, ''), IFNULL(new_value, ''), created_at
		FROM audit_log
		WHERE project_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.UserName, &e.ProjectID, &e.Entity, &e.EntityID,
			&e.Field, &e.OldValue, &e.NewValue, &e.CreatedAt); err != nil {
			log.Printf("❌ Ошибка чтения записи журнала: %v", err)
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// AuditFieldName — человекочитаемое название изменённого поля.
func AuditFieldName(field string) string {
	switch field {
	case "status":
		return "статус"
	case "budget":
		return "бюджет"
	case "assignee":
		return "исполнитель"
	default:
		return field
	}
}

// --- Аудируемые обёртки над операциями изменения ---
// Каждая обёртка читает прежнее значение, выполняет изменение и фиксирует его
// в журнале (только при фактическом отличии нового значения от старого).

// UpdateTaskStatusAudited меняет статус задачи и фиксирует изменение в журнале.
func UpdateTaskStatusAudited(actorID int64, actorName string, taskID int64, newStatus string) error {
	old, err := GetTaskByID(taskID)
	if err != nil {
		return err
	}
	if err := UpdateTaskStatus(taskID, newStatus); err != nil {
		return err
	}
	if old.Status != newStatus {
		AddAuditEntry(actorID, actorName, old.ProjectID, "task", taskID, "status",
			GetTaskStatusName(old.Status), GetTaskStatusName(newStatus))
	}
	return nil
}

// UpdateTaskAssigneeAudited назначает/снимает исполнителя и фиксирует изменение в журнале.
func UpdateTaskAssigneeAudited(actorID int64, actorName string, taskID int64, masterID *int64) error {
	old, err := GetTaskByID(taskID)
	if err != nil {
		return err
	}
	if err := UpdateTaskAssignee(taskID, masterID); err != nil {
		return err
	}
	oldName := GetMasterNameByID(old.AssigneeID)
	newName := GetMasterNameByID(masterID)
	if oldName != newName {
		AddAuditEntry(actorID, actorName, old.ProjectID, "task", taskID, "assignee", oldName, newName)
	}
	return nil
}

// UpdateProjectBudgetAudited обновляет бюджет проекта и фиксирует изменение в журнале.
// Прочие поля проекта (название, мастер) сохраняются без изменений.
func UpdateProjectBudgetAudited(actorID int64, actorName string, projectID int64, newBudget float64) error {
	old, err := GetProjectByID(projectID)
	if err != nil {
		return err
	}
	masterName := GetMasterNameByID(old.MasterID)
	if err := UpdateProject(projectID, old.Name, newBudget, masterName); err != nil {
		return err
	}
	if old.Budget != newBudget {
		AddAuditEntry(actorID, actorName, projectID, "project", projectID, "budget",
			fmt.Sprintf("%.2f", old.Budget), fmt.Sprintf("%.2f", newBudget))
	}
	return nil
}
