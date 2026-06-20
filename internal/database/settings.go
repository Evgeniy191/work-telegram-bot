package database

import (
	"database/sql"
	"log"
)

// GetUserSettings — возвращает настройки пользователя, создавая запись по умолчанию,
// если её ещё нет.
func GetUserSettings(userID int64) (UserSettings, error) {
	s := UserSettings{
		UserID:        userID,
		Lang:          "ru",
		Theme:         "light",
		Notifications: true,
	}

	err := DB.QueryRow(`
		SELECT user_id, lang, theme, notifications, IFNULL(quick_actions, '')
		FROM user_settings WHERE user_id = ?
	`, userID).Scan(&s.UserID, &s.Lang, &s.Theme, &s.Notifications, &s.QuickActions)

	if err == sql.ErrNoRows {
		// Создаём настройки по умолчанию
		if err := ensureUser(userID); err != nil {
			return s, err
		}
		_, err = DB.Exec(`
			INSERT INTO user_settings (user_id, lang, theme, notifications, updated_at)
			VALUES (?, 'ru', 'light', 1, datetime('now'))
		`, userID)
		if err != nil {
			log.Printf("❌ Ошибка создания настроек: %v", err)
			return s, err
		}
		return s, nil
	}
	if err != nil {
		log.Printf("❌ Ошибка чтения настроек: %v", err)
		return s, err
	}

	return s, nil
}

// SetNotifications — включает/выключает уведомления пользователя.
func SetNotifications(userID int64, on bool) error {
	// Гарантируем наличие записи настроек
	if _, err := GetUserSettings(userID); err != nil {
		return err
	}

	_, err := DB.Exec(`
		UPDATE user_settings SET notifications = ?, updated_at = datetime('now')
		WHERE user_id = ?
	`, on, userID)
	if err != nil {
		log.Printf("❌ Ошибка обновления уведомлений: %v", err)
	}
	return err
}

// ensureUser — добавляет пользователя в users, если его ещё нет (для FK настроек).
func ensureUser(userID int64) error {
	_, err := DB.Exec(`INSERT OR IGNORE INTO users (id, created_at) VALUES (?, datetime('now'))`, userID)
	return err
}
