package database

import (
	"strings"
)

// ProjectFilter — критерии выборки проектов. Пустые поля не ограничивают выборку.
type ProjectFilter struct {
	Status      string // статус проекта ("" — любой)
	MasterID    *int64 // мастер-исполнитель (nil — любой)
	OnlyOverdue bool   // только проекты с просроченными задачами
	Search      string // подстрока в названии/адресе/заказчике
	Sort        string // "created" (по умолчанию) | "name" | "budget"
}

// FilterProjects — выборка проектов пользователя по заданным критериям.
// Критерии комбинируются по «И»; сортировка задаётся полем Sort.
func FilterProjects(userID int64, f ProjectFilter) ([]Project, error) {
	where := []string{"p.user_id = ?"}
	args := []interface{}{userID}

	if f.Status != "" {
		where = append(where, "p.status = ?")
		args = append(args, f.Status)
	}
	if f.MasterID != nil {
		where = append(where, "p.master_id = ?")
		args = append(args, *f.MasterID)
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + s + "%"
		where = append(where, "(p.name LIKE ? OR IFNULL(p.address,'') LIKE ? OR IFNULL(p.customer,'') LIKE ?)")
		args = append(args, like, like, like)
	}
	if f.OnlyOverdue {
		where = append(where, `EXISTS (
			SELECT 1 FROM tasks t
			WHERE t.project_id = p.id
			  AND t.status <> 'completed'
			  AND t.deadline IS NOT NULL
			  AND t.deadline < datetime('now'))`)
	}

	order := "p.created_at DESC"
	switch f.Sort {
	case "name":
		order = "p.name COLLATE NOCASE ASC"
	case "budget":
		order = "p.budget DESC"
	}

	query := `SELECT ` + projectColumns + ` FROM projects p WHERE ` +
		strings.Join(where, " AND ") + ` ORDER BY ` + order

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProjects(rows), nil
}

// GetProjectStatuses — различные статусы проектов пользователя (для меню фильтра).
func GetProjectStatuses(userID int64) ([]string, error) {
	rows, err := DB.Query(
		`SELECT status, COUNT(*) FROM projects WHERE user_id = ? GROUP BY status ORDER BY COUNT(*) DESC, status ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []string
	for rows.Next() {
		var s string
		var n int
		if err := rows.Scan(&s, &n); err != nil {
			continue
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}

// GetProjectMasters — мастера, назначенные хотя бы на один проект пользователя
// (для меню фильтра по мастеру).
func GetProjectMasters(userID int64) ([]Master, error) {
	rows, err := DB.Query(`
		SELECT DISTINCT m.id, m.name, IFNULL(m.specialty, ''), m.active
		FROM masters m
		JOIN projects p ON p.master_id = m.id
		WHERE p.user_id = ?
		ORDER BY m.name COLLATE NOCASE ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var masters []Master
	for rows.Next() {
		var m Master
		if err := rows.Scan(&m.ID, &m.Name, &m.Specialty, &m.Active); err != nil {
			continue
		}
		masters = append(masters, m)
	}
	return masters, nil
}
