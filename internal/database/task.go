package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

// deadlineLayout — формат хранения дедлайна в БД (совместим с datetime('now')
// и парсингом при чтении). Хранение строкой делает сравнения дат предсказуемыми.
const deadlineLayout = "2006-01-02 15:04:05"

// deadlineParam приводит дедлайн к параметру запроса: строка нужного формата или NULL.
func deadlineParam(d *time.Time) interface{} {
	if d == nil {
		return nil
	}
	return d.Format(deadlineLayout)
}

// ParseDeadline разбирает дату из пользовательского ввода (ДД.ММ.ГГГГ или ГГГГ-ММ-ДД).
// Пустая строка → nil без ошибки (дедлайн не задан).
func ParseDeadline(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	for _, layout := range []string{"02.01.2006", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("неверный формат даты")
}

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
	`, projectID, name, description, deadlineParam(deadline))

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
		SELECT id, project_id, name, description, status, deadline, assignee_id, created_at, updated_at
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
		var deadline sql.NullTime
		var assignee sql.NullInt64
		rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Status, &deadline, &assignee, &t.CreatedAt, &t.UpdatedAt)

		if deadline.Valid {
			d := deadline.Time
			t.Deadline = &d
		}
		if assignee.Valid {
			a := assignee.Int64
			t.AssigneeID = &a
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetTaskByID — получает задачу по ID
func GetTaskByID(taskID int64) (*Task, error) {
	var t Task
	var deadline sql.NullTime
	var assignee sql.NullInt64

	err := DB.QueryRow(`
		SELECT id, project_id, name, description, status, deadline, assignee_id, created_at, updated_at
		FROM tasks WHERE id = ?
	`, taskID).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Description, &t.Status, &deadline, &assignee, &t.CreatedAt, &t.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("задача ID=%d не найдена", taskID)
	}
	if err != nil {
		return nil, err
	}

	if deadline.Valid {
		d := deadline.Time
		t.Deadline = &d
	}
	if assignee.Valid {
		a := assignee.Int64
		t.AssigneeID = &a
	}

	return &t, nil
}

// UpdateTaskAssignee — назначает (или снимает при nil) исполнителя задачи.
func UpdateTaskAssignee(taskID int64, masterID *int64) error {
	result, err := DB.Exec(
		"UPDATE tasks SET assignee_id = ?, updated_at = datetime('now') WHERE id = ?",
		masterID, taskID,
	)
	if err != nil {
		log.Printf("❌ Ошибка назначения исполнителя: %v", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("задача ID=%d не найдена", taskID)
	}
	return nil
}

// UpdateTaskStatus — меняет статус задачи
func UpdateTaskStatus(taskID int64, status string) error {
	validStatuses := map[string]bool{"pending": true, "in_progress": true, "review": true, "completed": true}
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
	case "review":
		return "🔎"
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
	case "review":
		return "На проверке"
	case "completed":
		return "Выполнена"
	default:
		return "Неизвестно"
	}
}

// CountProjectTasks — подсчитывает задачи проекта по статусам.
// IFNULL защищает от NULL, который SUM возвращает при отсутствии задач.
func CountProjectTasks(projectID int64) (total, pending, inProgress, completed int, err error) {
	err = DB.QueryRow(`
		SELECT
			COUNT(*) as total,
			IFNULL(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending,
			IFNULL(SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END), 0) as in_progress,
			IFNULL(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) as completed
		FROM tasks
		WHERE project_id = ?
	`, projectID).Scan(&total, &pending, &inProgress, &completed)

	return
}

// GetProjectProgress — возвращает прогресс проекта в процентах (0–100),
// рассчитанный как доля выполненных задач от общего числа.
func GetProjectProgress(projectID int64) (percent, total, completed int, err error) {
	total, _, _, completed, err = CountProjectTasks(projectID)
	if err != nil {
		return 0, 0, 0, err
	}
	percent = CalcProgress(total, completed)
	return percent, total, completed, nil
}

// CalcProgress — доля выполненного в процентах (с округлением).
func CalcProgress(total, completed int) int {
	if total <= 0 {
		return 0
	}
	return int(float64(completed)/float64(total)*100 + 0.5)
}

// PortfolioStats — сводная статистика по всем проектам пользователя.
type PortfolioStats struct {
	Projects    int
	TotalBudget float64
	Tasks       int
	Completed   int
	Overdue     int
	Progress    int
}

// GetPortfolioStats — агрегирует показатели по проектам пользователя для отчёта.
func GetPortfolioStats(userID int64) (PortfolioStats, error) {
	var s PortfolioStats

	// Проекты и бюджет — отдельным запросом, чтобы JOIN с задачами не множил бюджет.
	err := DB.QueryRow(`
		SELECT COUNT(*), IFNULL(SUM(budget), 0)
		FROM projects WHERE user_id = ?
	`, userID).Scan(&s.Projects, &s.TotalBudget)
	if err != nil {
		return s, err
	}

	// Задачи по всем проектам пользователя: всего, выполнено, просрочено.
	err = DB.QueryRow(`
		SELECT
			COUNT(t.id),
			IFNULL(SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END), 0),
			IFNULL(SUM(CASE WHEN t.status <> 'completed'
				AND t.deadline IS NOT NULL
				AND t.deadline < datetime('now') THEN 1 ELSE 0 END), 0)
		FROM tasks t
		JOIN projects p ON p.id = t.project_id
		WHERE p.user_id = ?
	`, userID).Scan(&s.Tasks, &s.Completed, &s.Overdue)
	if err != nil {
		return s, err
	}

	s.Progress = CalcProgress(s.Tasks, s.Completed)
	return s, nil
}

// IsTaskOverdue — задача просрочена, если есть дедлайн в прошлом и она не выполнена.
func IsTaskOverdue(t Task) bool {
	return t.Status != "completed" && t.Deadline != nil && t.Deadline.Before(time.Now())
}

// CountProjectOverdue — число просроченных (невыполненных с истёкшим сроком) задач проекта.
func CountProjectOverdue(projectID int64) (int, error) {
	var n int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM tasks
		WHERE project_id = ?
			AND status <> 'completed'
			AND deadline IS NOT NULL
			AND deadline < datetime('now')
	`, projectID).Scan(&n)
	return n, err
}

// UpdateTaskName — обновляет только название задачи
func UpdateTaskName(taskID int64, name string) error {
	_, err := DB.Exec("UPDATE tasks SET name = ?, updated_at = datetime('now') WHERE id = ?", name, taskID)
	return err
}

// UpdateTaskDescription — обновляет описание задачи
func UpdateTaskDescription(taskID int64, description string) error {
	_, err := DB.Exec("UPDATE tasks SET description = ?, updated_at = datetime('now') WHERE id = ?", description, taskID)
	return err
}

// UpdateTaskDeadline — обновляет дедлайн задачи (или сбрасывает, если nil)
func UpdateTaskDeadline(taskID int64, deadline *time.Time) error {
	_, err := DB.Exec(
		"UPDATE tasks SET deadline = ?, updated_at = datetime('now') WHERE id = ?",
		deadlineParam(deadline), taskID,
	)
	return err
}
