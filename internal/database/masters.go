package database

import (
	"database/sql"
	"fmt"
	"log"
)

// GetAllMasters — возвращает всех активных мастеров (для меню и клавиатур)
func GetAllMasters() ([]Master, error) {
	rows, err := DB.Query(`
		SELECT id, name, COALESCE(specialty, ''), active
		FROM masters
		WHERE active = 1
		ORDER BY id
	`)
	if err != nil {
		log.Printf("❌ Ошибка получения мастеров: %v", err)
		return nil, err
	}
	defer rows.Close()

	var masters []Master
	for rows.Next() {
		var m Master
		if err := rows.Scan(&m.ID, &m.Name, &m.Specialty, &m.Active); err != nil {
			log.Printf("❌ Ошибка чтения мастера: %v", err)
			continue
		}
		masters = append(masters, m)
	}

	return masters, nil
}

// GetMasterByID — возвращает мастера по ID
func GetMasterByID(masterID int64) (*Master, error) {
	var m Master
	err := DB.QueryRow(`
		SELECT id, name, COALESCE(specialty, ''), active
		FROM masters
		WHERE id = ?
	`, masterID).Scan(&m.ID, &m.Name, &m.Specialty, &m.Active)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("мастер ID=%d не найден", masterID)
	}
	if err != nil {
		log.Printf("❌ Ошибка получения мастера: %v", err)
		return nil, err
	}

	return &m, nil
}

// UpdateMasterName — обновляет ФИО мастера.
// Имя в таблице уникально, поэтому при дубликате вернётся понятная ошибка.
func UpdateMasterName(masterID int64, name string) error {
	// Проверяем, что новое имя не занято другим мастером
	var existingID int64
	err := DB.QueryRow("SELECT id FROM masters WHERE name = ? AND id != ?", name, masterID).Scan(&existingID)
	if err == nil {
		return fmt.Errorf("мастер с ФИО «%s» уже существует", name)
	} else if err != sql.ErrNoRows {
		return err
	}

	result, err := DB.Exec(
		"UPDATE masters SET name = ? WHERE id = ?",
		name, masterID,
	)
	if err != nil {
		log.Printf("❌ Ошибка обновления ФИО мастера: %v", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("мастер ID=%d не найден", masterID)
	}

	log.Printf("✅ ФИО мастера ID=%d обновлено → %s", masterID, name)
	return nil
}
