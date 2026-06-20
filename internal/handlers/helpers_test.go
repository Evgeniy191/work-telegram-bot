package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/Evgeniy191/work-telegram-bot/internal/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// TestMain поднимает временную БД (хендлеры обращаются к глобальному database.DB).
func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "h-test-*")
	if err != nil {
		panic(err)
	}
	cwd, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		panic(err)
	}
	// InitDB открывает соединение по относительному пути; после открытия
	// возвращаем cwd, чтобы coverage-профиль писался в исходную директорию.
	if err := database.InitDB(); err != nil {
		panic(err)
	}
	os.Chdir(cwd)

	code := m.Run()

	database.CloseDB()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// capture фиксирует исходящие запросы к (мок) Telegram API.
type capture struct {
	mu   sync.Mutex
	reqs []capReq
}

type capReq struct {
	method string
	form   url.Values
}

func (c *capture) add(method string, form url.Values) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reqs = append(c.reqs, capReq{method: method, form: form})
}

// joined объединяет тексты всех отправленных сообщений.
func (c *capture) joined() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var b strings.Builder
	for _, r := range c.reqs {
		b.WriteString(r.form.Get("text"))
		b.WriteString("\n")
	}
	return b.String()
}

// sends — число вызовов sendMessage.
func (c *capture) sends() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := 0
	for _, r := range c.reqs {
		if r.method == "sendMessage" {
			n++
		}
	}
	return n
}

func (c *capture) contains(sub string) bool {
	return strings.Contains(c.joined(), sub)
}

// newTestBot создаёт BotAPI, указывающий на локальный мок-сервер.
func newTestBot(t *testing.T) (*tgbotapi.BotAPI, *capture) {
	t.Helper()

	cap := &capture{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		method := parts[len(parts)-1]
		cap.add(method, r.Form)

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"ok":true,"result":{"message_id":1,"text":"ok","chat":{"id":1}}}`)
	}))
	t.Cleanup(srv.Close)

	bot, err := tgbotapi.NewBotAPIWithClient("123:abc", srv.URL+"/bot%s/%s", srv.Client())
	if err != nil {
		t.Fatalf("newTestBot: %v", err)
	}
	return bot, cap
}

func msgUpdate(chatID int64, text string) tgbotapi.Update {
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{UserName: "tester"},
			Text: text,
		},
	}
}

func cbUpdate(chatID int64, data string) tgbotapi.Update {
	return tgbotapi.Update{
		CallbackQuery: &tgbotapi.CallbackQuery{
			ID:      "cb1",
			From:    &tgbotapi.User{UserName: "tester"},
			Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}},
			Data:    data,
		},
	}
}

// dbCreateProjectWithTasks сидирует проект с total задачами, из которых
// completed помечены выполненными. Возвращает ID проекта.
func dbCreateProjectWithTasks(userID int64, total, completed int) (int64, error) {
	pid, err := database.CreateProject(userID, "Монтаж", "Seeded", 1000, "Не назначен")
	if err != nil {
		return 0, err
	}
	for i := 0; i < total; i++ {
		tid, err := database.CreateTask(pid, "Задача", "", nil)
		if err != nil {
			return 0, err
		}
		if i < completed {
			if err := database.UpdateTaskStatus(tid, "completed"); err != nil {
				return 0, err
			}
		}
	}
	return pid, nil
}
