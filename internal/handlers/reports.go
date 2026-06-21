package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
)

// buildProjectReportCSV собирает сводный отчёт по проекту в формате CSV
// (разделитель «;» и UTF-8 BOM — корректно открывается в Excel с кириллицей).
func buildProjectReportCSV(projectID int64) ([]byte, error) {
	p, err := database.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF") // UTF-8 BOM
	w := csv.NewWriter(&buf)
	w.Comma = ';'

	row := func(rec ...string) { w.Write(rec) }

	// Шапка проекта
	row("ПРОЕКТ", p.Name)
	row("Тип", p.Type)
	row("Статус", p.Status)
	row("Адрес", cardValueOrDash(p.Address))
	row("Заказчик", cardValueOrDash(p.Customer))
	row("Договор", cardValueOrDash(p.ContractNumber))
	row("Старт", cardDateOrDash(p.StartDate))
	row("Плановая сдача", cardDateOrDash(p.PlannedEnd))
	row("")

	// Бюджет план/факт
	bs, _ := database.GetProjectBudgetStatus(projectID)
	row("БЮДЖЕТ")
	row("План", fmt.Sprintf("%.2f", bs.Budget))
	row("Освоено", fmt.Sprintf("%.2f", bs.Spent))
	row("Остаток", fmt.Sprintf("%.2f", bs.Remaining))
	row("Освоение, %", strconv.Itoa(bs.Percent))
	row("")

	// Этапы
	row("ЭТАПЫ")
	row("Название", "Статус")
	stages, _ := database.GetProjectStages(projectID)
	for _, s := range stages {
		row(s.Name, database.GetTaskStatusName(s.Status))
	}
	row("")

	// Задачи
	row("ЗАДАЧИ")
	row("Название", "Статус", "Срок", "Исполнитель", "Просрочка")
	tasks, _ := database.GetProjectTasks(projectID)
	for _, t := range tasks {
		deadline := "—"
		if t.Deadline != nil {
			deadline = t.Deadline.Format("02.01.2006")
		}
		overdue := ""
		if database.IsTaskOverdue(t) {
			overdue = "просрочена"
		}
		row(t.Name, database.GetTaskStatusName(t.Status), deadline,
			database.GetMasterNameByID(t.AssigneeID), overdue)
	}
	row("")

	// Расходы
	row("РАСХОДЫ")
	row("Сумма", "Назначение", "Дата")
	expenses, _ := database.GetProjectExpenses(projectID)
	for _, e := range expenses {
		descr := e.Description
		if descr == "" {
			descr = "—"
		}
		row(fmt.Sprintf("%.2f", e.Amount), descr, e.CreatedAt.Format("02.01.2006"))
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
