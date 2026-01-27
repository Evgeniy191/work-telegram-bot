package database

import (
	"database/sql"
	"fmt"
	"log"
)

// CreateProject — добавляет новый проект в БД
func CreateProject(userID int64, projectType, name string, budget float64, masterName string) (int64, error) {
	// 1. Находим ID мастера по имени
	var masterID *int64 // Указатель, может быть NULL

	if masterName != "Не назначен" && masterName != "" {
		var id int64
		err := DB.QueryRow("SELECT id FROM masters WHERE name = ?", masterName).Scan(&id)
		if err == nil {
			masterID = &id // Сохраняем указатель на ID
		} else if err != sql.ErrNoRows {
			// Ошибка запроса (не "не найдено")
			log.Printf("❌ Ошибка поиска мастера: %v", err)
		}
		// Если sql.ErrNoRows — мастер не найден, masterID останется nil
	}

	// 2. Добавляем пользователя в users (если его ещё нет)
	_, err := DB.Exec(`
		INSERT OR IGNORE INTO users (id, created_at)
		VALUES (?, datetime('now'))
	`, userID)

	if err != nil {
		log.Printf("❌ Ошибка добавления пользователя: %v", err)
		return 0, err
	}

	// 3. Вставляем проект
	result, err := DB.Exec(`
		INSERT INTO projects (user_id, master_id, type, name, budget, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'В работе', datetime('now'), datetime('now'))
	`, userID, masterID, projectType, name, budget)

	if err != nil {
		log.Printf("❌ Ошибка создания проекта: %v", err)
		return 0, err
	}

	// 4. Получаем ID созданного проекта
	projectID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("✅ Проект ID=%d сохранён в БД (пользователь %d)", projectID, userID)
	return projectID, nil
}

// GetUserProjects — получает все проекты пользователя
func GetUserProjects(userID int64) ([]Project, error) {
	query := `
		SELECT p.id, p.user_id, p.master_id, p.type, p.name, 
		       p.budget, p.status, p.created_at, p.updated_at
		FROM projects p
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("❌ Ошибка получения проектов: %v", err)
		return nil, err
	}
	defer rows.Close()

	var projects []Project

	for rows.Next() {
		var p Project
		err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.MasterID, // Может быть NULL
			&p.Type,
			&p.Name,
			&p.Budget,
			&p.Status,
			&p.CreatedAt,
			&p.UpdatedAt,
		)

		if err != nil {
			log.Printf("❌ Ошибка чтения проекта: %v", err)
			continue
		}

		projects = append(projects, p)
	}

	log.Printf("📋 Найдено %d проектов для пользователя %d", len(projects), userID)
	return projects, nil
}

// GetProjectByID — получает проект по ID
func GetProjectByID(projectID int64) (*Project, error) {
	var p Project

	err := DB.QueryRow(`
		SELECT id, user_id, master_id, type, name, budget, status, created_at, updated_at
		FROM projects
		WHERE id = ?
	`, projectID).Scan(
		&p.ID,
		&p.UserID,
		&p.MasterID,
		&p.Type,
		&p.Name,
		&p.Budget,
		&p.Status,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("проект ID=%d не найден", projectID)
	}

	if err != nil {
		log.Printf("❌ Ошибка получения проекта: %v", err)
		return nil, err
	}

	return &p, nil
}

// UpdateProject — обновляет данные проекта
func UpdateProject(projectID int64, name string, budget float64, masterName string) error {
	// Находим ID мастера
	var masterID *int64

	if masterName != "Не назначен" && masterName != "" {
		var id int64
		err := DB.QueryRow("SELECT id FROM masters WHERE name = ?", masterName).Scan(&id)
		if err == nil {
			masterID = &id
		}
	}

	_, err := DB.Exec(`
		UPDATE projects 
		SET name = ?, budget = ?, master_id = ?, updated_at = datetime('now')
		WHERE id = ?
	`, name, budget, masterID, projectID)

	if err != nil {
		log.Printf("❌ Ошибка обновления проекта: %v", err)
		return err
	}

	log.Printf("✅ Проект ID=%d обновлён", projectID)
	return nil
}

// DeleteProject — удаляет проект
func DeleteProject(projectID int64) error {
	result, err := DB.Exec("DELETE FROM projects WHERE id = ?", projectID)

	if err != nil {
		log.Printf("❌ Ошибка удаления проекта: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		return fmt.Errorf("проект ID=%d не найден", projectID)
	}

	log.Printf("🗑️ Проект ID=%d удалён", projectID)
	return nil
}

// GetMasterNameByID — получает имя мастера по ID
func GetMasterNameByID(masterID *int64) string {
	if masterID == nil {
		return "Не назначен"
	}

	var name string
	err := DB.QueryRow("SELECT name FROM masters WHERE id = ?", *masterID).Scan(&name)

	if err != nil {
		return "Не назначен"
	}

	return name
}
