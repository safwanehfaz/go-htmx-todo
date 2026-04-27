package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Todo struct {
	ID        int
	Title     string
	Completed bool
}

type TodoStore struct {
	mu     sync.Mutex
	nextID int
	items  []Todo
}

func NewTodoStore() *TodoStore {
	return &TodoStore{nextID: 1}
}

func (s *TodoStore) Add(title string) Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := Todo{ID: s.nextID, Title: title}
	s.nextID++
	s.items = append(s.items, todo)
	return todo
}

func (s *TodoStore) Toggle(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Completed = !s.items[i].Completed
			return true
		}
	}
	return false
}

func (s *TodoStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.items {
		if s.items[i].ID == id {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return true
		}
	}
	return false
}

func (s *TodoStore) All() []Todo {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Todo, len(s.items))
	copy(out, s.items)
	return out
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
	store := NewTodoStore()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := listTmpl.ExecuteTemplate(w, "page", store.All()); err != nil {
			log.Printf("render page: %v", err)
		}
	})

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		title := strings.TrimSpace(r.FormValue("title"))
		if title != "" {
			store.Add(title)
		}
		renderList(w, store)
	})

	http.HandleFunc("/todos/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/todos/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 {
			http.NotFound(w, r)
			return
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		switch {
		case len(parts) == 2 && parts[1] == "toggle" && r.Method == http.MethodPost:
			store.Toggle(id)
			renderList(w, store)
			return
		case len(parts) == 1 && r.Method == http.MethodDelete:
			store.Delete(id)
			renderList(w, store)
			return
		default:
			http.NotFound(w, r)
			return
		}
	})

	log.Println("listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func renderList(w http.ResponseWriter, store *TodoStore) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := listTmpl.ExecuteTemplate(w, "list", store.All()); err != nil {
		log.Printf("render list: %v", err)
	}
}
