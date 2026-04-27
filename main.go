package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	appName     = "togx"
	defaultHost = "127.0.0.1"
	defaultPort = "8080"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type persistedTodos struct {
	NextID int    `json:"next_id"`
	Items  []Todo `json:"items"`
}

type TodoStore struct {
	mu       sync.Mutex
	nextID   int
	items    []Todo
	dataFile string
}

func NewTodoStore(dataFile string) (*TodoStore, error) {
	s := &TodoStore{nextID: 1, dataFile: dataFile}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *TodoStore) Add(title string) (Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{ID: s.nextID, Title: title}
	s.nextID++
	s.items = append(s.items, todo)
	if err := s.saveLocked(); err != nil {
		return Todo{}, err
	}
	return todo, nil
}

func (s *TodoStore) Toggle(id int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Completed = !s.items[i].Completed
			if err := s.saveLocked(); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *TodoStore) Delete(id int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			if err := s.saveLocked(); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, nil
}

func (s *TodoStore) All() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Todo, len(s.items))
	copy(out, s.items)
	return out
}

func (s *TodoStore) load() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	var p persistedTodos
	if err := json.Unmarshal(data, &p); err != nil {
		return err
	}
	s.items = p.Items
	if p.NextID > 0 {
		s.nextID = p.NextID
		return nil
	}
	maxID := 0
	for _, t := range s.items {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	s.nextID = maxID + 1
	return nil
}

func (s *TodoStore) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.dataFile), 0o755); err != nil {
		return err
	}
	p := persistedTodos{NextID: s.nextID, Items: s.items}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.dataFile + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.dataFile)
}

var pageTmpl = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Go + HTMX Todo</title>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <style>
    :root {
      --bg: #f6f4ef;
      --card: #fffdf8;
      --text: #1f2933;
      --muted: #6b7280;
      --line: #e5dfd4;
      --accent: #0f766e;
      --accent-hover: #0b5f58;
      --danger: #b91c1c;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Iosevka Aile", "JetBrains Mono", monospace;
      color: var(--text);
      background: radial-gradient(circle at top right, #f0ebe0 0, var(--bg) 45%);
      min-height: 100vh;
      display: grid;
      place-items: center;
      padding: 1rem;
    }
    .app {
      width: min(760px, 100%);
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 14px;
      padding: 1.2rem;
      box-shadow: 0 18px 45px rgba(31, 41, 51, 0.08);
    }
    h1 {
      margin: 0 0 1rem;
      font-size: clamp(1.4rem, 4vw, 1.9rem);
      letter-spacing: -0.02em;
    }
    .new-todo {
      display: grid;
      grid-template-columns: 1fr auto;
      gap: .6rem;
      margin-bottom: 1rem;
    }
    input[type="text"] {
      width: 100%;
      border: 1px solid var(--line);
      border-radius: 10px;
      background: #fff;
      font: inherit;
      padding: .7rem .85rem;
    }
    button {
      border: 0;
      border-radius: 10px;
      background: var(--accent);
      color: #fff;
      padding: .7rem .95rem;
      font: inherit;
      cursor: pointer;
      transition: background-color 140ms ease;
    }
    button:hover { background: var(--accent-hover); }
    ul {
      list-style: none;
      margin: 0;
      padding: 0;
      border-top: 1px solid var(--line);
    }
    li {
      display: grid;
      grid-template-columns: 1fr auto;
      align-items: center;
      gap: .75rem;
      padding: .8rem 0;
      border-bottom: 1px solid var(--line);
      animation: reveal 160ms ease-out;
    }
    .todo-main {
      display: flex;
      align-items: center;
      gap: .65rem;
      min-width: 0;
    }
    .todo-title {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .done .todo-title {
      text-decoration: line-through;
      color: var(--muted);
    }
    .delete {
      background: transparent;
      color: var(--danger);
      border: 1px solid #f3c7c7;
      padding: .45rem .7rem;
    }
    .delete:hover {
      background: #fee2e2;
    }
    .empty {
      color: var(--muted);
      padding: 1rem 0 .4rem;
      text-align: center;
    }
    @keyframes reveal {
      from { opacity: 0; transform: translateY(6px); }
      to { opacity: 1; transform: translateY(0); }
    }
    @media (max-width: 520px) {
      .new-todo { grid-template-columns: 1fr; }
      button { width: 100%; }
      li { grid-template-columns: 1fr; }
      .delete { width: 100%; }
    }
  </style>
</head>
<body>
  <main class="app">
    <h1>Todo App</h1>

    <form class="new-todo"
          hx-post="/todos"
          hx-target="#todo-list"
          hx-swap="outerHTML"
          hx-on::after-request="if(event.detail.successful) this.reset()">
      <input type="text" name="title" placeholder="Write a task..." required maxlength="120" />
      <button type="submit">Add</button>
    </form>

    {{template "list" .}}
  </main>
</body>
</html>`))

var listTmpl = template.Must(template.Must(pageTmpl.Clone()).New("list").Parse(`<ul id="todo-list">
  {{if .}}
    {{range .}}
      <li class="{{if .Completed}}done{{end}}">
        <div class="todo-main">
          <input type="checkbox"
                 {{if .Completed}}checked{{end}}
                 hx-post="/todos/{{.ID}}/toggle"
                 hx-target="#todo-list"
                 hx-swap="outerHTML" />
          <span class="todo-title">{{.Title}}</span>
        </div>
        <button class="delete"
                hx-delete="/todos/{{.ID}}"
                hx-target="#todo-list"
                hx-swap="outerHTML">
          Delete
        </button>
      </li>
    {{end}}
  {{else}}
    <li class="empty">No tasks yet.</li>
  {{end}}
</ul>`))

func main() {
	if err := run(os.Args[1:]); err != nil {
		logError(err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return runStart([]string{"--foreground"})
	}

	switch args[0] {
	case "start":
		return runStart(args[1:])
	case "stop", "quit":
		return runStop(args[1:])
	case "status":
		return runStatus()
	case "autostart":
		return runAutostart(args[1:])
	default:
		return fmt.Errorf("unknown command %q\n\nusage: %s [start|stop|quit|status|autostart]", args[0], appName)
	}
}

func runStart(args []string) error {
	fs := flag.NewFlagSet("start", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	foreground := fs.Bool("foreground", false, "run in foreground")
	detach := fs.Bool("detach", false, "run in background")
	host := fs.String("host", defaultHost, "host to bind")
	port := fs.String("port", defaultPort, "port to bind")
	if err := fs.Parse(args); err != nil {
		return err
	}

	runInForeground := *foreground || !*detach
	if !runInForeground {
		if err := startDetached(*host, *port); err != nil {
			return err
		}
		logSuccess("started in background")
		return nil
	}

	return runServer(*host, *port)
}

func startDetached(host, port string) error {
	p, err := readPIDFile()
	if err == nil && processExists(p) {
		return fmt.Errorf("already running with pid %d", p)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "start", "--foreground", "--host", host, "--port", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("failed to start in background: %w", err)
		}
		return fmt.Errorf("failed to start in background")
	case <-time.After(700 * time.Millisecond):
	}

	if !processExists(cmd.Process.Pid) {
		return fmt.Errorf("failed to start in background")
	}
	return nil
}

func runStop(args []string) error {
	fs := flag.NewFlagSet("stop", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	force := fs.Bool("force", false, "force kill")
	forceShort := fs.Bool("f", false, "force kill")
	if err := fs.Parse(args); err != nil {
		return err
	}

	pid, err := readPIDFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logWarn("not running")
			return nil
		}
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	useForce := *force || *forceShort
	if useForce {
		if err := process.Kill(); err != nil {
			return err
		}
		_ = removePIDFile()
		logWarn("force stopped")
		return nil
	}

	if err := process.Signal(os.Interrupt); err != nil {
		if err := process.Kill(); err != nil {
			return err
		}
	}

	for i := 0; i < 60; i++ {
		time.Sleep(100 * time.Millisecond)
		if !processExists(pid) {
			_ = removePIDFile()
			logSuccess("stopped")
			return nil
		}
	}

	return fmt.Errorf("could not stop pid %d gracefully, use --force", pid)
}

func runStatus() error {
	pid, err := readPIDFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logWarn("status: stopped")
			return nil
		}
		return err
	}

	if processExists(pid) {
		logInfo(fmt.Sprintf("status: running (pid %d)", pid))
		return nil
	}
	_ = removePIDFile()
	logWarn("status: stopped (stale pid cleaned)")
	return nil
}

func runAutostart(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: %s autostart [enable|disable|status]", appName)
	}

	if runtime.GOOS != "linux" {
		return fmt.Errorf("autostart is currently supported on linux with systemd --user")
	}

	svcPath, err := autostartServicePath()
	if err != nil {
		return err
	}

	switch args[0] {
	case "enable":
		exe, err := os.Executable()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(svcPath), 0o755); err != nil {
			return err
		}
		svc := fmt.Sprintf("[Unit]\nDescription=togx todo app\nAfter=network.target\n\n[Service]\nType=simple\nExecStart=%s start --foreground\nRestart=on-failure\nRestartSec=2\n\n[Install]\nWantedBy=default.target\n", exe)
		if err := os.WriteFile(svcPath, []byte(svc), 0o644); err != nil {
			return err
		}
		if err := runCmd("systemctl", "--user", "daemon-reload"); err != nil {
			return err
		}
		if err := runCmd("systemctl", "--user", "enable", "--now", "togx.service"); err != nil {
			return err
		}
		logSuccess("autostart enabled")
		return nil
	case "disable":
		_ = runCmd("systemctl", "--user", "disable", "--now", "togx.service")
		_ = os.Remove(svcPath)
		_ = runCmd("systemctl", "--user", "daemon-reload")
		logSuccess("autostart disabled")
		return nil
	case "status":
		err := runCmd("systemctl", "--user", "is-enabled", "togx.service")
		if err != nil {
			logWarn("autostart: disabled")
			return nil
		}
		logSuccess("autostart: enabled")
		return nil
	default:
		return fmt.Errorf("usage: %s autostart [enable|disable|status]", appName)
	}
}

func runServer(host, port string) error {
	pid, err := readPIDFile()
	if err == nil && processExists(pid) {
		return fmt.Errorf("already running with pid %d", pid)
	}

	if err := writePIDFile(os.Getpid()); err != nil {
		return err
	}
	defer func() {
		_ = removePIDFile()
	}()

	dataFile, err := todoDataFile()
	if err != nil {
		return err
	}
	store, err := NewTodoStore(dataFile)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := listTmpl.ExecuteTemplate(w, "page", store.All()); err != nil {
			logError("render page failed")
		}
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))
		if title != "" {
			if _, err := store.Add(title); err != nil {
				http.Error(w, "save failed", http.StatusInternalServerError)
				return
			}
			logInfo("todo added")
		}
		renderList(w, store)
	})

	mux.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		id, action, ok := parseTodoAction(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		switch {
		case action == "toggle" && r.Method == http.MethodPost:
			_, err := store.Toggle(id)
			if err != nil {
				http.Error(w, "save failed", http.StatusInternalServerError)
				return
			}
			logInfo("todo toggled")
			renderList(w, store)
			return
		case action == "delete" && r.Method == http.MethodDelete:
			_, err := store.Delete(id)
			if err != nil {
				http.Error(w, "save failed", http.StatusInternalServerError)
				return
			}
			logInfo("todo deleted")
			renderList(w, store)
			return
		default:
			http.NotFound(w, r)
			return
		}
	})

	srv := &http.Server{
		Addr:              host + ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		<-sigCh
		logWarn("shutdown signal received")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	logSuccess(fmt.Sprintf("server running at http://%s:%s", host, port))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logSuccess("server stopped")
	return nil
}

func parseTodoAction(path string) (int, string, bool) {
	trimmed := strings.TrimPrefix(path, "/todos/")
	parts := strings.Split(trimmed, "/")
	if len(parts) < 1 {
		return 0, "", false
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", false
	}
	if len(parts) == 1 {
		return id, "delete", true
	}
	if len(parts) == 2 && parts[1] == "toggle" {
		return id, "toggle", true
	}
	return 0, "", false
}

func renderList(w http.ResponseWriter, store *TodoStore) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := listTmpl.ExecuteTemplate(w, "list", store.All()); err != nil {
		logError("render list failed")
	}
}

func todoDataFile() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, appName, "todos.json"), nil
}

func pidFilePath() string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir != "" {
		return filepath.Join(runtimeDir, appName+".pid")
	}
	return filepath.Join(os.TempDir(), appName+".pid")
}

func readPIDFile() (int, error) {
	b, err := os.ReadFile(pidFilePath())
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func writePIDFile(pid int) error {
	return os.WriteFile(pidFilePath(), []byte(strconv.Itoa(pid)), 0o644)
}

func removePIDFile() error {
	err := os.Remove(pidFilePath())
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func processExists(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = p.Signal(syscall.Signal(0))
	return err == nil
}

func autostartServicePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "systemd", "user", "togx.service"), nil
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func colorEnabled() bool {
	term := os.Getenv("TERM")
	if term == "" || term == "dumb" {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func colorize(code, msg string) string {
	if !colorEnabled() {
		return msg
	}
	return "\033[" + code + "m" + msg + "\033[0m"
}

func logInfo(msg string) {
	fmt.Printf("%s %s\n", colorize("36", "[INFO]"), msg)
}

func logWarn(msg string) {
	fmt.Printf("%s %s\n", colorize("33", "[WARN]"), msg)
}

func logError(msg string) {
	fmt.Printf("%s %s\n", colorize("31", "[ERR ]"), msg)
}

func logSuccess(msg string) {
	fmt.Printf("%s %s\n", colorize("32", "[ OK ]"), msg)
}
