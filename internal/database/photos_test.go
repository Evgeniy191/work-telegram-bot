package database

import "testing"

func TestTaskPhotosCRUD(t *testing.T) {
	defer setupTestDB(t)()

	pid, _ := CreateProject(900, "Монтаж", "P", 100, "Не назначен")
	tid, _ := CreateTask(pid, "Задача", "", nil)

	// Пусто
	if n, err := CountTaskPhotos(tid); err != nil || n != 0 {
		t.Fatalf("CountTaskPhotos(пусто) = %d, %v", n, err)
	}
	if ph, err := GetTaskPhotos(tid); err != nil || len(ph) != 0 {
		t.Fatalf("GetTaskPhotos(пусто) = %d, %v", len(ph), err)
	}

	// Добавляем два фото
	if _, err := AddTaskPhoto(tid, "file_1", "до работ"); err != nil {
		t.Fatalf("AddTaskPhoto: %v", err)
	}
	if _, err := AddTaskPhoto(tid, "file_2", ""); err != nil {
		t.Fatalf("AddTaskPhoto: %v", err)
	}

	n, _ := CountTaskPhotos(tid)
	if n != 2 {
		t.Errorf("CountTaskPhotos = %d, want 2", n)
	}

	photos, _ := GetTaskPhotos(tid)
	if len(photos) != 2 {
		t.Fatalf("GetTaskPhotos = %d, want 2", len(photos))
	}
	if photos[0].FileID != "file_1" || photos[0].Caption != "до работ" {
		t.Errorf("первое фото неверно: %+v", photos[0])
	}
	if photos[1].FileID != "file_2" {
		t.Errorf("второе фото неверно: %+v", photos[1])
	}
}

func TestTaskPhotosError(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()
	DB.Close()

	if _, err := AddTaskPhoto(1, "f", ""); err == nil {
		t.Error("AddTaskPhoto на закрытой БД должен вернуть ошибку")
	}
	if _, err := GetTaskPhotos(1); err == nil {
		t.Error("GetTaskPhotos на закрытой БД должен вернуть ошибку")
	}
	if _, err := CountTaskPhotos(1); err == nil {
		t.Error("CountTaskPhotos на закрытой БД должен вернуть ошибку")
	}
}
