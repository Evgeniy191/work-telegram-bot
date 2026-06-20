package database

import "testing"

func TestUpdateTaskAssignee(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(800, "Монтаж", "P", 100, "Не назначен")
	tid, _ := CreateTask(pid, "Задача", "", nil)

	masters, _ := GetAllMasters()
	if len(masters) == 0 {
		t.Fatal("нет сидированных мастеров")
	}
	mID := masters[0].ID

	// Назначаем
	if err := UpdateTaskAssignee(tid, &mID); err != nil {
		t.Fatalf("UpdateTaskAssignee: %v", err)
	}
	task, _ := GetTaskByID(tid)
	if task.AssigneeID == nil || *task.AssigneeID != mID {
		t.Errorf("исполнитель не назначен: %v", task.AssigneeID)
	}

	// Проверяем чтение через список проекта
	tasks, _ := GetProjectTasks(pid)
	if len(tasks) != 1 || tasks[0].AssigneeID == nil || *tasks[0].AssigneeID != mID {
		t.Error("исполнитель не читается в GetProjectTasks")
	}

	// Снимаем
	if err := UpdateTaskAssignee(tid, nil); err != nil {
		t.Fatalf("UpdateTaskAssignee(nil): %v", err)
	}
	task, _ = GetTaskByID(tid)
	if task.AssigneeID != nil {
		t.Error("исполнитель должен быть снят")
	}

	// Несуществующая задача
	if err := UpdateTaskAssignee(999999, &mID); err == nil {
		t.Error("ожидалась ошибка для несуществующей задачи")
	}
}

func TestUpdateTaskAssigneeError(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	id := int64(1)
	if err := UpdateTaskAssignee(1, &id); err == nil {
		t.Error("ожидалась ошибка на закрытой БД")
	}
}
