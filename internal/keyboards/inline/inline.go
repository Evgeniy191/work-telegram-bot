package inline

import (
	"log"

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
