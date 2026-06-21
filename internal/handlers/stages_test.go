package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestCallbackStagesMenu(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7500)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 100000, "")

	// Пустое меню этапов
	cap := cb(t, chatID, fmt.Sprintf("stages_%d", pid), mgr)
	if !cap.contains("Этапы проекта") {
		t.Errorf("ожидалось меню этапов, получили: %s", cap.joined())
	}

	// Ошибки
	cap = cb(t, chatID, "stages_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "stages_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackSeedAndStageCycle(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7501)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")

	// Типовые этапы
	cap := cb(t, chatID, fmt.Sprintf("seed_stages_%d", pid), mgr)
	if !cap.contains("Прогресс по этапам") {
		t.Errorf("ожидался прогресс после посева, получили: %s", cap.joined())
	}
	stages, _ := database.GetProjectStages(pid)
	if len(stages) != len(database.StandardStages) {
		t.Fatalf("ожидалось %d этапов, получили %d", len(database.StandardStages), len(stages))
	}

	// Цикл статуса этапа: pending → in_progress
	cb(t, chatID, fmt.Sprintf("stage_next_%d", stages[0].ID), mgr)
	st, _ := database.GetStageByID(stages[0].ID)
	if st.Status != "in_progress" {
		t.Errorf("после переключения статус = %q, want in_progress", st.Status)
	}

	// Ошибки
	cap = cb(t, chatID, "stage_next_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "stage_next_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}

func TestCallbackAddStageFlow(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7502)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")

	// Старт добавления
	cb(t, chatID, fmt.Sprintf("add_stage_%d", pid), mgr)
	if mgr.GetState(chatID) != fsm.StateAddingStageName {
		t.Error("add_stage не установил состояние")
	}
	if mgr.GetData(chatID).EditingProjectID != pid {
		t.Error("EditingProjectID не сохранён")
	}

	// Пустое имя
	bot, cap := newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "   "), mgr)
	if !cap.contains("пустым") {
		t.Errorf("ожидалась ошибка пустого имени, получили: %s", cap.joined())
	}

	// Корректное имя → этап создан
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "Электрика"), mgr)
	if !cap.contains("Этап добавлен") {
		t.Errorf("ожидалось подтверждение, получили: %s", cap.joined())
	}
	stages, _ := database.GetProjectStages(pid)
	if len(stages) != 1 || stages[0].Name != "Электрика" {
		t.Errorf("этап не сохранён: %+v", stages)
	}

	// Ошибка ID
	cap = cb(t, chatID, "add_stage_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
}

func TestMyProjectsShowsStageProgress(t *testing.T) {
	bot, cap := newTestBot(t)
	chatID := int64(7503)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом", 1, "")
	s1, _ := database.AddStage(pid, "Фундамент")
	database.AddStage(pid, "Стены")
	database.UpdateStageStatus(s1, "completed")

	HandleMyProjects(bot, msgUpdate(chatID, "📋 Проекты"))
	if !cap.contains("Этапы: 1/2") {
		t.Errorf("ожидался прогресс по этапам в карточке, получили: %s", cap.joined())
	}
}
