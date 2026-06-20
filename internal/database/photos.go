package database

import "log"

// AddTaskPhoto — сохраняет фотофиксацию (Telegram file_id) к задаче.
func AddTaskPhoto(taskID int64, fileID, caption string) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO task_photos (task_id, file_id, caption, created_at)
		VALUES (?, ?, ?, datetime('now'))
	`, taskID, fileID, caption)
	if err != nil {
		log.Printf("❌ Ошибка сохранения фото: %v", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	log.Printf("📷 Фото ID=%d добавлено к задаче %d", id, taskID)
	return id, nil
}

// GetTaskPhotos — список фотофиксаций задачи (старые первыми).
func GetTaskPhotos(taskID int64) ([]TaskPhoto, error) {
	rows, err := DB.Query(`
		SELECT id, task_id, file_id, IFNULL(caption, ''), created_at
		FROM task_photos WHERE task_id = ?
		ORDER BY created_at ASC, id ASC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []TaskPhoto
	for rows.Next() {
		var p TaskPhoto
		if err := rows.Scan(&p.ID, &p.TaskID, &p.FileID, &p.Caption, &p.CreatedAt); err != nil {
			log.Printf("❌ Ошибка чтения фото: %v", err)
			continue
		}
		photos = append(photos, p)
	}
	return photos, nil
}

// CountTaskPhotos — число фотофиксаций задачи.
func CountTaskPhotos(taskID int64) (int, error) {
	var n int
	err := DB.QueryRow("SELECT COUNT(*) FROM task_photos WHERE task_id = ?", taskID).Scan(&n)
	return n, err
}
