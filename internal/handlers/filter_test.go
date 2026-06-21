package handlers

import (
	"fmt"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestProjectFilterCallbacks(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(8100)

	// Башня — мастер Иванов Иван, просроченная задача; Домик — без мастера/просрочки.
	p1, _ := database.CreateProject(chatID, "Монтаж", "Башня", 5000, "Иванов Иван")
	to, _ := database.CreateTask(p1, "t", "", nil)
	past := past72h()
	database.UpdateTaskDeadline(to, &past)
	database.CreateProject(chatID, "Ремонт", "Домик", 1000, "Не назначен")

	// Меню фильтра.
	if cap := cb(t, chatID, "pfilter", mgr); !cap.contains("Фильтр и сортировка") {
		t.Errorf("ожидалось меню фильтра: %s", cap.joined())
	}

	// Все проекты.
	if cap := cb(t, chatID, "pf_all", mgr); !cap.contains("найдено: 2") {
		t.Errorf("ожидалось 2 проекта: %s", cap.joined())
	}

	// Только с просрочкой → Башня.
	cap := cb(t, chatID, "pf_overdue", mgr)
	if !cap.contains("найдено: 1") || !cap.contains("Башня") {
		t.Errorf("ожидался 1 просроченный (Башня): %s", cap.joined())
	}

	// Сортировки.
	if cap := cb(t, chatID, "pf_sort_name", mgr); !cap.contains("по названию") {
		t.Errorf("ожидалась сортировка по названию: %s", cap.joined())
	}
	if cap := cb(t, chatID, "pf_sort_budget", mgr); !cap.contains("по бюджету") {
		t.Errorf("ожидалась сортировка по бюджету: %s", cap.joined())
	}

	// Подменю и фильтр по статусу.
	if cap := cb(t, chatID, "pf_status", mgr); !cap.contains("статус") {
		t.Errorf("ожидалось подменю статусов: %s", cap.joined())
	}
	if cap := cb(t, chatID, "pf_st_В работе", mgr); !cap.contains("найдено: 2") {
		t.Errorf("ожидалось 2 по статусу «В работе»: %s", cap.joined())
	}

	// Подменю и фильтр по мастеру.
	if cap := cb(t, chatID, "pf_master", mgr); !cap.contains("мастера") {
		t.Errorf("ожидалось подменю мастеров: %s", cap.joined())
	}
	masters, _ := database.GetProjectMasters(chatID)
	if len(masters) == 0 {
		t.Fatal("ожидался хотя бы один мастер на проектах")
	}
	if cap := cb(t, chatID, fmt.Sprintf("pf_ms_%d", masters[0].ID), mgr); !cap.contains("Башня") {
		t.Errorf("ожидался проект мастера (Башня): %s", cap.joined())
	}
	if cap := cb(t, chatID, "pf_ms_abc", mgr); !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID мастера: %s", cap.joined())
	}

	// Поиск запускает FSM-диалог.
	cb(t, chatID, "pf_search", mgr)
	if mgr.GetState(chatID) != fsm.StateSearchingProjects {
		t.Errorf("ожидалось состояние поиска, получили: %s", mgr.GetState(chatID))
	}
}

func TestProjectSearchFSM(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(8101)
	database.CreateProject(chatID, "Монтаж", "Северный мост", 100, "Не назначен")
	database.CreateProject(chatID, "Ремонт", "Южная дорога", 200, "Не назначен")

	mgr.SetState(chatID, fsm.StateSearchingProjects)
	bot, cap := newTestBot(t)
	if !HandleFSMMessage(bot, msgUpdate(chatID, "мост"), mgr) {
		t.Fatal("FSM не обработал поиск")
	}
	if !cap.contains("Северный мост") || cap.contains("Южная дорога") {
		t.Errorf("ожидался только «Северный мост»: %s", cap.joined())
	}
	if mgr.GetState(chatID) != fsm.StateIdle {
		t.Error("состояние должно сброситься после поиска")
	}

	// Пустой запрос — поиск отменяется.
	mgr.SetState(chatID, fsm.StateSearchingProjects)
	bot, cap = newTestBot(t)
	HandleFSMMessage(bot, msgUpdate(chatID, "   "), mgr)
	if !cap.contains("Пустой запрос") {
		t.Errorf("ожидалась отмена при пустом запросе: %s", cap.joined())
	}
}
