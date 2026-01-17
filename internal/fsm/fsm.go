package fsm

import (
	"log"
	"sync"
)

// State — тип состояния пользователя
type State string

// Константы состояний
const (
	StateIdle                  State = "idle"                    // Обычный режим
	StateCreatingProject       State = "creating_project"        // Создание проекта: ждём название
	StateCreatingProjectBudget State = "creating_project_budget" // Ждём бюджет
	StateCreatingTask          State = "creating_task"           // Создание задачи
	StateEditingProject        State = "editing_project"         // Редактирование
)

// UserData — данные пользователя в процессе диалога
type UserData struct {
	ProjectName   string
	ProjectBudget string
	TaskName      string
}

// Manager — менеджер состояний FSM
type Manager struct {
	mu     sync.RWMutex
	states map[int64]State     // chatID -> текущее состояние
	data   map[int64]*UserData // chatID -> временные данные
}

// NewManager — создаёт новый FSM менеджер
func NewManager() *Manager {
	return &Manager{
		states: make(map[int64]State),
		data:   make(map[int64]*UserData),
	}
}

// GetState — получить текущее состояние пользователя
func (m *Manager) GetState(chatID int64) State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[chatID]
	if !exists {
		return StateIdle // По умолчанию
	}
	return state
}

// SetState — установить состояние пользователя
func (m *Manager) SetState(chatID int64, state State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[chatID] = state
	log.Printf("🔄 FSM: Пользователь %d → состояние '%s'", chatID, state)
}

// ResetState — вернуть в обычный режим
func (m *Manager) ResetState(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, chatID)
	delete(m.data, chatID)
	log.Printf("🔄 FSM: Пользователь %d → сброс состояния", chatID)
}

// GetData — получить временные данные пользователя
func (m *Manager) GetData(chatID int64) *UserData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.data[chatID]
	if !exists {
		return &UserData{} // Пустые данные
	}
	return data
}

// SetData — сохранить временные данные
func (m *Manager) SetData(chatID int64, data *UserData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[chatID] = data
	log.Printf("💾 FSM: Данные пользователя %d обновлены", chatID)
}
