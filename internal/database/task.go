package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// CreateTask — добавляет новую задачу к проекту
func CreateTask(projectID int64, name, description string, deadline *time.Time) (int64, error) {
	// Проверяем проект
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", projectID).Scan(&exists)
	if err != nil || !exists {
		return 0, fmt.Errorf("проект ID=%d не найден", projectID)
	}

	// Вставляем задачу
	result, err := DB.Exec(`
		INSERT INTO tasks (project_id, name, description, status, deadline, created_at, updated_at)
		VALUES (?, ?, ?, 'pending', ?, datetime('now'), datetime('now'))
	`, projectID, name, description, deadline)

	if err != nil {
		log.Printf("❌ Ошибка создания задачи: %v", err)
		return 0, err
	}

	taskID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("✅ Задача ID=%d создана для проекта %d", taskID, projectID)
	return taskID, nil
}

// GetProjectTasks — получает задачи проекта
func GetProjectTasks(projectID int64) ([]Task, error) {
	query := `
		SELECT id, project_id, name, description, status, deadline, created_at, updated_at
		FROM tasks WHERE project_id = ?
		ORDER BY 
			CASE status WHEN 'pending' THEN 1 WHEN 'in_progress' THEN 2 WHEN 'completed' THEN 3 END,
			created_at ASC
	`

	rows, err := DB.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var deadlineStr sql.NullString
		rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Status, &deadlineStr, &t.CreatedAt, &t.UpdatedAt)

		if deadlineStr.Valid {
			deadline, _ := time.Parse("2006-01-02 15:04:05", deadlineStr.String)
			t.Deadline = &deadline
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetTaskByID — получает задачу по ID
func GetTaskByID(taskID int64) (*Task, error) {
	var t Task
	var deadlineStr sql.NullString

	err := DB.QueryRow(`
		SELECT id, project_id, name, description, status, deadline, created_at, updated_at
		FROM tasks WHERE id = ?
	`, taskID).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Status, &deadlineStr, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("задача ID=%d не найдена", taskID)
	}
	if err != nil {
		return nil, err
	}

	if deadlineStr.Valid {
		deadline, _ := time.Parse("2006-01-02 15:04:05", deadlineStr.String)
		t.Deadline = &deadline
	}

	return &t, nil
}

// UpdateTaskStatus — меняет статус задачи
func UpdateTaskStatus(taskID int64, status string) error {
	validStatuses := map[string]bool{"pending": true, "in_progress": true, "completed": true}
	if !validStatuses[status] {
		return fmt.Errorf("неверный статус: %s", status)
	}

	result, err := DB.Exec("UPDATE tasks SET status = ?, updated_at = datetime('now') WHERE id = ?", status, taskID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("задача ID=%d не найдена", taskID)
	}

	log.Printf("✅ Статус задачи %d → %s", taskID, status)
	return nil
}

// UpdateTask — обновляет данные задачи
func UpdateTask(taskID int64, name, description string) error {
	result, err := DB.Exec(`
		UPDATE tasks 
		SET name = ?, description = ?, updated_at = datetime('now')
		WHERE id = ?
	`, name, description, taskID)

	if err != nil {
		log.Printf("❌ Ошибка обновления задачи: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("задача ID=%d не найдена", taskID)
	}

	log.Printf("✅ Задача ID=%d обновлена", taskID)
	return nil
}

// DeleteTask — удаляет задачу
func DeleteTask(taskID int64) error {
	result, err := DB.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("задача ID=%d не найдена", taskID)
	}

	log.Printf("🗑️ Задача ID=%d удалена", taskID)
	return nil
}

// GetTaskStatusEmoji — возвращает эмодзи для статуса
func GetTaskStatusEmoji(status string) string {
	switch status {
	case "pending":
		return "🕐"
	case "in_progress":
		return "⚙️"
	case "completed":
		return "✅"
	default:
		return "❓"
	}
}

// GetTaskStatusName — возвращает русское название статуса
func GetTaskStatusName(status string) string {
	switch status {
	case "pending":
		return "Ожидает"
	case "in_progress":
		return "В процессе"
	case "completed":
		return "Выполнена"
	default:
		return "Неизвестно"
	}
}

// CountProjectTasks — подсчитывает задачи проекта по статусам
func CountProjectTasks(projectID int64) (total, pending, inProgress, completed int, err error) {
	err = DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks
		WHERE project_id = ?
	`, projectID).Scan(&total, &pending, &inProgress, &completed)

	return
}
