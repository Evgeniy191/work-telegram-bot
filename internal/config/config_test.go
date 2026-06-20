package config

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestLoadWithToken(t *testing.T) {
	t.Setenv("TELEGRAM_TOKEN", "123:abc")

	cfg := Load()
	if cfg == nil {
		t.Fatal("Load() вернул nil")
	}
	if cfg.TelegramToken != "123:abc" {
		t.Errorf("TelegramToken = %q, want %q", cfg.TelegramToken, "123:abc")
	}
	if !cfg.Debug {
		t.Error("Debug должен быть true")
	}
}

// TestLoadMissingToken проверяет ветку log.Fatal через подпроцесс,
// т.к. log.Fatal вызывает os.Exit.
func TestLoadMissingToken(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		Load() // должен вызвать os.Exit(1)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestLoadMissingToken")
	// Чистое окружение без TELEGRAM_TOKEN
	env := []string{"BE_CRASHER=1"}
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "TELEGRAM_TOKEN=") && !strings.HasPrefix(e, "BE_CRASHER=") {
			env = append(env, e)
		}
	}
	cmd.Env = env

	err := cmd.Run()
	if exitErr, ok := err.(*exec.ExitError); ok && !exitErr.Success() {
		return // ожидаемый аварийный выход
	}
	t.Fatalf("ожидался выход с ошибкой, получили err=%v", err)
}
