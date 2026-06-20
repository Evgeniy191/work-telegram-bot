package keyboards

import (
	"fmt"
	"log"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ProjectsList — inline-клавиатура выбора проектов
func ProjectsList() tgbotapi.InlineKeyboardMarkup {
	log.Println("📋 Генерация inline-клавиатуры 'Проекты'")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏗️ Монтаж труб А", "project_1"),
			tgbotapi.NewInlineKeyboardButtonData("🔧 Оборудование Б", "project_2"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔩 Монтаж м/к", "project_3"),
			tgbotapi.NewInlineKeyboardButtonData("➕ Новый проект", "project_new"),
		),
	)
}

// TasksList — inline-клавиатура задач проекта
func TasksList() tgbotapi.InlineKeyboardMarkup {
	log.Println("📋 Генерация inline-клавиатуры 'Задачи'")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Проверить документы", "task_docs"),
			tgbotapi.NewInlineKeyboardButtonData("⏳ Закупка материалов", "task_materials"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Отчёт", "task_report"),
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "back_projects"),
		),
	)
}

// ConfirmAction — подтверждение действия
func ConfirmAction() tgbotapi.InlineKeyboardMarkup {
	log.Println("📋 Генерация inline-клавиатуры 'Подтверждение'")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить", "confirm_yes"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "confirm_no"),
		),
	)
}

// ProjectTypeKeyboard — inline-клавиатура выбора типа проекта
func ProjectTypeKeyboard() tgbotapi.InlineKeyboardMarkup {
	log.Println("📋 Генерация inline-клавиатуры 'Тип проекта'")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Монтаж", "type_montazh"),
			tgbotapi.NewInlineKeyboardButtonData("🛠️ Ремонт", "type_remont"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⚙️ Установка", "type_ustanovka"),
			tgbotapi.NewInlineKeyboardButtonData("🏗️ Строительство", "type_stroitelstvo"),
		),
	)
}

// BackButton — создаёт inline-кнопку "Назад"
func BackButton(callbackData string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", callbackData),
		),
	)
}

// MastersKeyboard — inline-клавиатура выбора мастера.
// Список строится динамически из БД, поэтому отредактированные ФИО
// сразу появляются здесь, а выбор идёт по ID (callback master_pick_<id>).
func MastersKeyboard() tgbotapi.InlineKeyboardMarkup {
	log.Println("📋 Генерация inline-клавиатуры 'Мастера'")

	masters, err := database.GetAllMasters()
	if err != nil {
		log.Printf("❌ Не удалось загрузить мастеров: %v", err)
	}

	var rows [][]tgbotapi.InlineKeyboardButton

	// По два мастера в ряд
	for i := 0; i < len(masters); i += 2 {
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(
				"👷 "+masters[i].Name,
				fmt.Sprintf("master_pick_%d", masters[i].ID),
			),
		}
		if i+1 < len(masters) {
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(
				"👷 "+masters[i+1].Name,
				fmt.Sprintf("master_pick_%d", masters[i+1].ID),
			))
		}
		rows = append(rows, row)
	}

	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Без мастера", "master_none"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", "back_to_budget"),
		),
	)

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// NotificationsKeyboard — переключатель уведомлений.
func NotificationsKeyboard(on bool) tgbotapi.InlineKeyboardMarkup {
	label := "🔔 Включить уведомления"
	if on {
		label = "🔕 Выключить уведомления"
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, "toggle_notifications"),
		),
	)
}

// MastersManageKeyboard — inline-клавиатура управления мастерами.
// Для каждого мастера кнопка правки ФИО (callback edit_master_name_<id>).
func MastersManageKeyboard() tgbotapi.InlineKeyboardMarkup {
	masters, err := database.GetAllMasters()
	if err != nil {
		log.Printf("❌ Не удалось загрузить мастеров: %v", err)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, m := range masters {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ "+m.Name,
				fmt.Sprintf("edit_master_name_%d", m.ID),
			),
		))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Главное меню", "back_menu"),
	))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
