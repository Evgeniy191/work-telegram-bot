package database

import "testing"

func TestReviewStatus(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(1400, "Монтаж", "P", 1, "Не назначен")
	tid, _ := CreateTask(pid, "T", "", nil)

	if err := UpdateTaskStatus(tid, "review"); err != nil {
		t.Fatalf("UpdateTaskStatus(review): %v", err)
	}
	task, _ := GetTaskByID(tid)
	if task.Status != "review" {
		t.Errorf("статус = %q, want review", task.Status)
	}

	if GetTaskStatusEmoji("review") != "🔎" {
		t.Error("неверный эмодзи для review")
	}
	if GetTaskStatusName("review") != "На проверке" {
		t.Error("неверное имя для review")
	}
}
