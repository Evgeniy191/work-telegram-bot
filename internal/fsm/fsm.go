package fsm

import (
	"log"
	"sync"
)

// State представляет состояние FSM
type State string

// Константы состояний
const (
	StateIdle                    State = "idle"
	StateCreatingProjectType     State = "creating_project_type"
	StateCreatingProject         State = "creating_project"
	StateCreatingProjectBudget   State = "creating_project_budget"
	StateCreatingProjectMaster   State = "creating_project_master"
	StateCreatingTask            State = "creating_task"
	StateEditingProject          State = "editing_project"
	StateEditingProjectName      State = "editing_project_name"
	StateEditingProjectBudget    State = "editing_project_budget"
	StateCreatingTaskName        State = "creating_task_name"
	StateCreatingTaskDescription State = "creating_task_description"
	StateCreatingTaskDeadline    State = "creating_task_deadline"
	StateEditingTaskName         State = "editing_task_name"
	StateEditingTaskDescription  State = "editing_task_description"
	StateEditingTaskDeadline     State = "editing_task_deadline"
	StateEditingMasterName       State = "editing_master_name"
	StateAddingTaskPhoto         State = "adding_task_photo"
	StateAddingExpenseAmount     State = "adding_expense_amount"
	StateAddingExpenseDesc       State = "adding_expense_desc"
)

// UserData хранит временные данные пользователя
type UserData struct {
	ProjectType      string
	ProjectName      string
	ProjectBudget    string
	ProjectMaster    string
	TaskName         string
	EditingProjectID int64
	TaskProjectID    int64
	TaskDescription  string
	EditingTaskID    int64
	EditingMasterID  int64
	ExpenseAmount    float64
}

// Manager управляет состояниями пользователей
type Manager struct {
	mu     sync.RWMutex
	states map[int64]State
	data   map[int64]*UserData
}

// NewManager создаёт новый FSM Manager
func NewManager() *Manager {
	return &Manager{
		states: make(map[int64]State),
		data:   make(map[int64]*UserData),
	}
}

// GetState возвращает состояние пользователя
func (m *Manager) GetState(chatID int64) State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[chatID]
	if !exists {
		return StateIdle
	}
	return state
}

// SetState устанавливает состояние пользователя
func (m *Manager) SetState(chatID int64, state State) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[chatID] = state
	log.Printf("🔄 FSM: Пользователь %d → состояние '%s'", chatID, state)
}

// GetData возвращает данные пользователя
func (m *Manager) GetData(chatID int64) *UserData {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, exists := m.data[chatID]
	if !exists {
		data = &UserData{}
		m.data[chatID] = data
	}
	return data
}

// SetData устанавливает данные пользователя
func (m *Manager) SetData(chatID int64, data *UserData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[chatID] = data
	log.Printf("💾 FSM: Данные пользователя %d обновлены", chatID)
}

// ResetState сбрасывает состояние и данные
func (m *Manager) ResetState(chatID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, chatID)
	delete(m.data, chatID)
	log.Printf("🔄 FSM: Пользователь %d → сброс состояния", chatID)
}
