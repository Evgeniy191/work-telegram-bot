package database

import (
	"fmt"
	"log"
)

// AddMaterial — добавляет заявку на материал к проекту.
func AddMaterial(projectID int64, name, quantity string) (int64, error) {
	var exists bool
	if err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM projects WHERE id = ?)", projectID).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("проект ID=%d не найден", projectID)
	}

	result, err := DB.Exec(`
		INSERT INTO materials (project_id, name, quantity, status, created_at)
		VALUES (?, ?, ?, 'ordered', datetime('now'))
	`, projectID, name, quantity)
	if err != nil {
		log.Printf("❌ Ошибка добавления материала: %v", err)
		return 0, err
	}
	return result.LastInsertId()
}

// GetProjectMaterials — заявки на материалы проекта.
func GetProjectMaterials(projectID int64) ([]Material, error) {
	rows, err := DB.Query(`
		SELECT id, project_id, name, IFNULL(quantity, ''), status, created_at
		FROM materials WHERE project_id = ?
		ORDER BY created_at ASC, id ASC
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []Material
	for rows.Next() {
		var m Material
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Name, &m.Quantity, &m.Status, &m.CreatedAt); err != nil {
			log.Printf("❌ Ошибка чтения материала: %v", err)
			continue
		}
		materials = append(materials, m)
	}
	return materials, nil
}

// GetMaterialByID — материал по ID.
func GetMaterialByID(materialID int64) (*Material, error) {
	var m Material
	err := DB.QueryRow(`
		SELECT id, project_id, name, IFNULL(quantity, ''), status, created_at
		FROM materials WHERE id = ?
	`, materialID).Scan(&m.ID, &m.ProjectID, &m.Name, &m.Quantity, &m.Status, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("материал ID=%d не найден", materialID)
	}
	return &m, nil
}

// UpdateMaterialStatus — меняет статус поставки материала.
func UpdateMaterialStatus(materialID int64, status string) error {
	valid := map[string]bool{"ordered": true, "in_transit": true, "received": true}
	if !valid[status] {
		return fmt.Errorf("неверный статус материала: %s", status)
	}
	result, err := DB.Exec("UPDATE materials SET status = ? WHERE id = ?", status, materialID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return fmt.Errorf("материал ID=%d не найден", materialID)
	}
	return nil
}

// NextMaterialStatus — цикл статуса: ordered → in_transit → received → ordered.
func NextMaterialStatus(status string) string {
	switch status {
	case "ordered":
		return "in_transit"
	case "in_transit":
		return "received"
	default:
		return "ordered"
	}
}

// MaterialStatusEmoji — значок статуса поставки.
func MaterialStatusEmoji(status string) string {
	switch status {
	case "ordered":
		return "🛒"
	case "in_transit":
		return "🚚"
	case "received":
		return "✅"
	default:
		return "❓"
	}
}

// MaterialStatusName — название статуса поставки.
func MaterialStatusName(status string) string {
	switch status {
	case "ordered":
		return "Заказано"
	case "in_transit":
		return "В пути"
	case "received":
		return "Получено"
	default:
		return "Неизвестно"
	}
}
