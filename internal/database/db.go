package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB — глобальная переменная для подключения к БД
var DB *sql.DB

// InitDB — инициализация базы данных
func InitDB() error {
	// 1. Создаём папку data (если её нет)
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	// 2. Путь к файлу БД
	dbPath := filepath.Join(dataDir, "bot.db")
	log.Printf("📂 База данных: %s", dbPath)

	// 3. Открываем соединение
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 4. Проверяем соединение
	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("✅ Соединение с БД установлено")

	// 5. Создаём таблицы
	if err = createTables(); err != nil {
		return err
	}

	log.Println("✅ Таблицы успешно созданы")
	return nil
}

// createTables — создание всех таблиц
func createTables() error {
	query := `
	-- 1. Users
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		first_name TEXT,
		premium BOOLEAN DEFAULT 0,
		role TEXT DEFAULT 'manager',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- 2. Masters
	CREATE TABLE IF NOT EXISTS masters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		specialty TEXT,
		active BOOLEAN DEFAULT 1
	);

	-- Заполняем мастеров
	INSERT OR IGNORE INTO masters (name, specialty) VALUES
	('Иванов Иван', 'Монтаж'),
	('Петров Пётр', 'Ремонт'),
	('Сидоров Сергей', 'Установка'),
	('Кузнецов Андрей', 'Строительство');

	-- 3. Projects
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		master_id INTEGER,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		budget REAL NOT NULL CHECK (budget >= 0),
		status TEXT DEFAULT 'В работе',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (master_id) REFERENCES masters(id)
	);

	CREATE INDEX IF NOT EXISTS idx_projects_user ON projects(user_id);
	CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

	-- 4. Tasks
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		deadline DATETIME,
		assignee_id INTEGER,
		status TEXT DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);

	-- 4b. Task photos (фотофиксация работ)
	CREATE TABLE IF NOT EXISTS task_photos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		file_id TEXT NOT NULL,
		caption TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_task_photos_task ON task_photos(task_id);

	-- 4c. Expenses (учёт фактических расходов по проекту)
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		amount REAL NOT NULL CHECK (amount >= 0),
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_expenses_project ON expenses(project_id);

	-- 5. Reports
	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER,
		type TEXT NOT NULL,
		file_path TEXT,
		generated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	-- 6. User Settings
	CREATE TABLE IF NOT EXISTS user_settings (
		user_id INTEGER PRIMARY KEY,
		lang TEXT DEFAULT 'ru',
		theme TEXT DEFAULT 'light',
		notifications BOOLEAN DEFAULT 1,
		quick_actions TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		log.Printf("❌ Ошибка создания таблиц: %v", err)
		return err
	}

	migrateSchema()
	return nil
}

// migrateSchema — безопасные миграции для уже существующих БД.
// ALTER упадёт с «duplicate column» на свежей БД (колонка уже в CREATE) —
// это ожидаемо и игнорируется.
func migrateSchema() {
	DB.Exec("ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'manager'")
}

// CloseDB — закрытие соединения
func CloseDB() error {
	if DB != nil {
		log.Println("🔒 Закрываем соединение с БД")
		return DB.Close()
	}
	return nil
}
