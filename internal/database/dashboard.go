package database

import "log"

// StatusCount — число проектов в конкретном статусе (для разбивки в дашборде).
type StatusCount struct {
	Status string
	Count  int
}

// PortfolioDashboard — расширенная панель показателей портфеля пользователя:
// базовые метрики (см. PortfolioStats) плюс освоение бюджета и разбивка
// проектов по статусам.
type PortfolioDashboard struct {
	PortfolioStats               // встроенные метрики: проекты, бюджет, задачи, прогресс
	Spent          float64       // освоено по бюджету (сумма расходов всех проектов)
	Remaining      float64       // остаток бюджета (может быть отрицательным при перерасходе)
	BudgetPercent  int           // процент освоения бюджета
	OverBudget     bool          // признак перерасхода
	StatusCounts   []StatusCount // разбивка проектов по статусам
}

// GetPortfolioDashboard — собирает полную панель показателей портфеля пользователя.
// Переиспользует GetPortfolioStats для базовых метрик и добавляет освоение бюджета
// и разбивку проектов по статусам.
func GetPortfolioDashboard(userID int64) (PortfolioDashboard, error) {
	var d PortfolioDashboard

	stats, err := GetPortfolioStats(userID)
	if err != nil {
		return d, err
	}
	d.PortfolioStats = stats

	// Освоение бюджета: сумма расходов по всем проектам пользователя.
	err = DB.QueryRow(`
		SELECT IFNULL(SUM(e.amount), 0)
		FROM expenses e
		JOIN projects p ON p.id = e.project_id
		WHERE p.user_id = ?
	`, userID).Scan(&d.Spent)
	if err != nil {
		return d, err
	}
	d.Remaining = d.TotalBudget - d.Spent
	d.OverBudget = d.Spent > d.TotalBudget
	if d.TotalBudget > 0 {
		d.BudgetPercent = int(d.Spent/d.TotalBudget*100 + 0.5)
	}

	// Разбивка проектов по статусам (по убыванию количества).
	rows, err := DB.Query(`
		SELECT status, COUNT(*)
		FROM projects
		WHERE user_id = ?
		GROUP BY status
		ORDER BY COUNT(*) DESC, status ASC
	`, userID)
	if err != nil {
		return d, err
	}
	defer rows.Close()

	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
			log.Printf("❌ Ошибка чтения статуса проекта: %v", err)
			continue
		}
		d.StatusCounts = append(d.StatusCounts, sc)
	}

	return d, nil
}
