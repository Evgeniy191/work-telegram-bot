package database

import "time"

// Project — модель проекта
type Project struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	MasterID  *int64    `json:"master_id"` // Указатель (может быть NULL)
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Budget    float64   `json:"budget"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Task — модель задачи
type Task struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Master — модель мастера
type Master struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
	Active    bool   `json:"active"`
}

// UserSettings — настройки пользователя
type UserSettings struct {
	UserID        int64     `json:"user_id"`
	Lang          string    `json:"lang"`
	Theme         string    `json:"theme"`
	Notifications bool      `json:"notifications"`
	QuickActions  string    `json:"quick_actions"` // JSON строка
	UpdatedAt     time.Time `json:"updated_at"`
}
