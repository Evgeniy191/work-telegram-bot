package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	"github.com/Evgeniy191/work-telegram-bot/internal/fsm"
)

func TestBuildProjectReportCSV(t *testing.T) {
	chatID := int64(7800)
	pid, _ := database.CreateProject(chatID, "Строительство", "Дом у реки", 100000, "")
	database.UpdateProjectCardText(pid, "customer", "ООО Заказчик")
	tid, _ := database.CreateTask(pid, "Фундамент залить", "", nil)
	database.UpdateTaskStatus(tid, "review")
	database.AddStage(pid, "Фундамент")
	database.AddExpense(pid, 5000, "бетон")

	data, err := buildProjectReportCSV(pid)
	if err != nil {
		t.Fatalf("buildProjectReportCSV: %v", err)
	}
	csv := string(data)

	if !strings.HasPrefix(csv, "\xEF\xBB\xBF") {
		t.Error("ожидался UTF-8 BOM в начале CSV")
	}
	for _, want := range []string{"Дом у реки", "ООО Заказчик", "ЗАДАЧИ", "Фундамент залить", "На проверке", "РАСХОДЫ", "бетон", "ЭТАПЫ"} {
		if !strings.Contains(csv, want) {
			t.Errorf("в отчёте нет %q\n%s", want, csv)
		}
	}
}

func TestBuildProjectReportCSVMissing(t *testing.T) {
	if _, err := buildProjectReportCSV(999999); err == nil {
		t.Error("ожидалась ошибка для несуществующего проекта")
	}
}

func TestCallbackExport(t *testing.T) {
	mgr := fsm.NewManager()
	chatID := int64(7801)
	pid, _ := database.CreateProject(chatID, "Монтаж", "P", 1, "")

	cap := cb(t, chatID, fmt.Sprintf("export_%d", pid), mgr)
	if cap.method("sendDocument") != 1 {
		t.Errorf("ожидалась отправка документа, методы: %d", cap.method("sendDocument"))
	}

	cap = cb(t, chatID, "export_abc", mgr)
	if !cap.contains("Неверный ID") {
		t.Errorf("ожидалась ошибка ID, получили: %s", cap.joined())
	}
	cap = cb(t, chatID, "export_999999", mgr)
	if !cap.contains("не найден") {
		t.Errorf("ожидалась ошибка 'не найден', получили: %s", cap.joined())
	}
}
