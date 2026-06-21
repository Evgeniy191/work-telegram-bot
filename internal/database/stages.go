package database

import (
	"fmt"
	"log"
)

// StandardStages — типовые этапы строительного проекта.
var StandardStages = []string{"Фундамент", "Стены", "Кровля", "Отделка"}

// AddStage — добавляет этап в конец списка проекта.
func AddStage(projectID int64, name string) (int64, error) {
	var exists bool
	if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", projectID).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("проект ID=%d не найден", projectID)
	}

	var pos int
	DB.QueryRow("SELECT IFNULL(MAX(position), 0) FROM stages WHERE project_id = ?", projectID).Scan(&pos)

	result, err := DB.Exec(`
		INSERT INTO stages (project_id, name, position, status, created_at)
		VALUES (?, ?, ?, 'pending', datetime('now'))
	`, projectID, name, pos+1)
	if err != nil {
		log.Printf("❌ Ошибка добавления этапа: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

// SeedStandardStages — добавляет типовые этапы стройпроекта.
func SeedStandardStages(projectID int64) error {
	for _, name := range StandardStages {
		if _, err := AddStage(projectID, name); err != nil {
			return err
		}
	}
	return nil
}

// GetProjectStages — этапы проекта по порядку.
func GetProjectStages(projectID int64) ([]Stage, error) {
	rows, err := DB.Query(`
		SELECT id, project_id, name, position, status, created_at
		FROM stages WHERE project_id = ?
		ORDER BY position ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stages []Stage
	for rows.Next() {
		var s Stage
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Name, &s.Position, &s.Status, &s.CreatedAt); err != nil {
			log.Printf("❌ Ошибка чтения этапа: %v", err)
			continue
		}
		stages = append(stages, s)
	}
	return stages, nil
}

// GetStageByID — этап по ID.
func GetStageByID(stageID int64) (*Stage, error) {
	var s Stage
	err := DB.QueryRow(`
		SELECT id, project_id, name, position, status, created_at
		FROM stages WHERE id = ?
	`, stageID).Scan(&s.ID, &s.ProjectID, &s.Name, &s.Position, &s.Status, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("этап ID=%d не найден", stageID)
	}
	return &s, nil
}

// UpdateStageStatus — меняет статус этапа.
func UpdateStageStatus(stageID int64, status string) error {
	valid := map[string]bool{"pending": true, "in_progress": true, "completed": true}
	if !valid[status] {
		return fmt.Errorf("неверный статус этапа: %s", status)
	}
	result, err := DB.Exec("UPDATE stages SET status = ? WHERE id = ?", status, stageID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return fmt.Errorf("этап ID=%d не найден", stageID)
	}
	return nil
}

// NextStageStatus — следующий статус по циклу pending → in_progress → completed → pending.
func NextStageStatus(status string) string {
	switch status {
	case "pending":
		return "in_progress"
	case "in_progress":
		return "completed"
	default:
		return "pending"
	}
}

// GetProjectStageProgress — прогресс проекта по этапам (percent, total, completed).
func GetProjectStageProgress(projectID int64) (percent, total, completed int, err error) {
	err = DB.QueryRow(`
		SELECT
			COUNT(*),
			IFNULL(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0)
		FROM stages WHERE project_id = ?
	`, projectID).Scan(&total, &completed)
	if err != nil {
		return 0, 0, 0, err
	}
	percent = CalcProgress(total, completed)
	return percent, total, completed, nil
}
