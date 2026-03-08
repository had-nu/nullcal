# nullcal — Project Specification

> A local web calendar and task manager. Google Calendar as a replaceable backend.  
> Your attention is the interface. Everything else is noise.

*Spec version: 2.0 — March 2026*  
*Reflects implementation state after Phase 1c. Next review: after Phase 1d (GCal OAuth).*

---

## 1. Philosophy

Most calendar tools are built around the assumption that your schedule manages you.  
nullcal inverts this: you define intent, the tool reflects it.

Three invariants hold across all development phases:

1. **Local-first.** All data lives in SQLite. Network is optional, never required.
2. **Backend-agnostic.** Google Calendar is Adapter v1. CalDAV, iCal, or nothing are equally valid.
3. **LLM as suggestion, never authority.** When Ollama is introduced, it proposes — you decide.

---

## 2. Scope

### MVP (Phase 1)
- Local web server serving a dashboard with week calendar, to-do list, kanban board
- Layout splits controlled by the web UI
- Local task management (create, edit, complete, delete, move status)
- Fixed routine blocks (e.g. Monday = Vexil, Tuesday = Wardex) from config
- Google Calendar read/write via OAuth2
- Conflict detection between local tasks and GCal events

### Phase 2
- Full CRUD for GCal events from within the UI
- Configurable notification system (terminal bell / desktop via `notify-send`)
- CalDAV adapter (replaces GCal adapter, same interface)

### Phase 3 — Intent Dispatcher (Ollama)
- Tag-based routing: task tagged `[wardex]` → Ollama receives project context
- LLM returns: suggested subtasks + proposed calendar blocks
- User reviews and accepts/rejects each suggestion explicitly
- Mechanical operations (creating events, writing to DB) only execute after human confirmation

---

## 3. Architecture

```
nullcal/
├── cmd/
│   └── root.go                  ← Cobra entrypoint; serve subcommand
├── internal/
│   ├── config/                  ← XDG config, RoutineBlock loader, OAuth token path
│   ├── store/
│   │   ├── types.go             ← Task, CalendarEvent, TaskStatus, Recurrence
│   │   ├── task.go              ← CRUD + SetTaskStatus + query helpers
│   │   ├── store.go             ← SQLite connection, WAL, embedded migrations
│   │   └── migrations/
│   │       └── 001_initial.sql  ← tasks, calendar_events, sync_state, schema_migrations
│   └── web/
│       ├── hub.go               ← WebSocket hub and state broadcast
│       ├── client.go            ← WebSocket client connection
│       ├── handler.go           ← HTTP handlers for index.html and static assets
│       └── static/              ← HTML, CSS, JS frontend assets
├── pkg/
│   └── timeutil/                ← WeekBounds, DaysOfWeek, WeekNumber (ISO 8601)
├── Makefile
└── lazygo.yml
```

### Layer Responsibilities

| Layer | Owns | Does NOT own |
|---|---|---|
| `store` | All persistence, SQLite schema, migrations, status transitions | Business logic, sync decisions |
| `sync/gcal` | GCal API calls, OAuth2 token lifecycle | Local state, conflict resolution |
| `web` | Local HTTP server, WebSocket broadcasts, handling UI updates | Direct DB access, API calls |
| `intent` | Ollama prompt construction, response parsing | Executing any calendar/task operations |

---

## 4. Core Types

### store.Task

```go
type Task struct {
    ID          string
    Title       string
    Description string
    ProjectTag  string     // e.g. "wardex", "vexil", "namesniper"
    Status      TaskStatus // "backlog" | "doing" | "done"
    DueAt       *time.Time
    CompletedAt *time.Time
    Recurrence  Recurrence // "" | "daily" | "weekly" | "monthly"
    CreatedAt   time.Time
}
```

`SetTaskStatus` automatically sets `CompletedAt` on transition to `done` and clears it otherwise.

### config.RoutineBlock

```go
type RoutineBlock struct {
    Weekday    time.Weekday
    StartTime  string // "09:00"
    EndTime    string // "11:00"
    Label      string // "Wardex — feature dev"
    ProjectTag string
}
```

Defined once in `config.yaml`, rendered on every matching weekday. Not persisted in SQLite.

### store.CalendarEvent

```go
type CalendarEvent struct {
    ExternalID  string
    Source      string // "gcal" | "caldav"
    Title       string
    StartAt     time.Time
    EndAt       time.Time
    Description string
    SyncedAt    time.Time
}
```

### CalendarAdapter interface (defined at consumer site)

```go
type CalendarAdapter interface {
    ListEvents(ctx context.Context, from, to time.Time) ([]CalendarEvent, error)
    CreateEvent(ctx context.Context, e CalendarEvent) (CalendarEvent, error)
    UpdateEvent(ctx context.Context, e CalendarEvent) error
    DeleteEvent(ctx context.Context, externalID string) error
}
```

---

## 5. Google Calendar OAuth2 Flow

OAuth2 tokens are stored at `$XDG_CONFIG_HOME/nullcal/token.json`.  
On first run, nullcal opens the browser for consent and writes the token.  
Subsequent runs refresh silently using the stored refresh token.

The token file **must** be in `.gitignore`.

```
$XDG_CONFIG_HOME/nullcal/
├── config.yaml      ← user preferences, routine blocks, adapter selection
└── token.json       ← OAuth2 token (never committed)
```

---

## 6. SQLite Schema

```sql
CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    project_tag  TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'backlog'
                     CHECK(status IN ('backlog', 'doing', 'done')),
    due_at       DATETIME,
    completed_at DATETIME,
    recurrence   TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE calendar_events (
    external_id  TEXT PRIMARY KEY,
    source       TEXT NOT NULL,
    title        TEXT NOT NULL,
    start_at     DATETIME NOT NULL,
    end_at       DATETIME NOT NULL,
    description  TEXT,
    synced_at    DATETIME NOT NULL
);

CREATE TABLE sync_state (
    adapter      TEXT PRIMARY KEY,
    last_sync_at DATETIME,
    sync_token   TEXT   -- GCal incremental sync token
);

CREATE TABLE schema_migrations (
    version    TEXT PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

WAL mode and foreign keys enforced on connection open. Migrations embedded in binary via `//go:embed migrations/*.sql`. Each migration runs in its own transaction. `ListTasks()` returns all tasks ordered by `created_at DESC` — no date-range filtering at the store layer.

---

## 7. Dependency Stack

| Dependency | Purpose | Rationale |
|---|---|---|
| `github.com/mattn/go-sqlite3` | SQLite driver | Proven, WAL-compatible |
| `github.com/google/uuid` | Task ID generation | RFC 4122 UUIDs |
| `google.golang.org/api/calendar/v3` | GCal API client | Official SDK (Phase 1d) |
| `golang.org/x/oauth2` | OAuth2 token management | stdlib-tier quality (Phase 1d) |
| `github.com/spf13/cobra` | CLI entrypoint | Consistent with Wardex |
| `golang.org/x/net/websocket` | Web UI | Live bidirectional updates |

No ORM. Raw SQL via `database/sql`. Queries are readable and auditable.

---

## 8. Phase 3 — Intent Dispatcher (Design Preview)

```
[user creates task: "implement VEX reachability in Wardex"]
         │
         ▼
intent.Classify(task) → ProjectTag: "wardex", Intent: "feature"
         │
         ▼
ollama.Prompt(context{project: wardex, task: title, history: recent_tasks})
         │
         ▼
LLM returns: suggested subtasks + calendar block proposal
         │
         ▼
Web UI renders proposal panel → user accepts / rejects each item
         │
         ▼
store.CreateTask() / gcal.CreateEvent()  ← only on explicit accept
```

The LLM never writes to the store or calls any adapter directly.  
**Narrative only, no mechanical authority** — same invariant as Arcanum.

---

## 9. Development Phases & Milestones

| Phase | Milestone | Status |
|---|---|---|
| 0 | Scaffold via lazy.go, go.mod, SQLite connection, config loader | ✅ Done |
| 1a | Store layer: CRUD for tasks, status transitions, embedded migrations | ✅ Done |
| 1b | Web: Static week view, calendar render, routine blocks | ✅ Done |
| 1c | Web: Split-pane layout model, todo/kanban views | ✅ Done |
| 1d | GCal adapter: OAuth2 flow + ListEvents | ⏳ Next |
| 1e | Sync: write CalendarEvents to local store, render in week view | — |
| 1f | GCal write: CreateEvent from UI | — |
| 2a | CalDAV adapter (same interface as gcal) | — |
| 2b | Notification layer | — |
| 3 | Intent dispatcher + Ollama integration | TBD |

---

## 10. Relation to the Ecosystem

```
nullcal (this project)
    │
    │  routine block: [wardex] → opens Wardex context
    │  routine block: [vexil]  → opens Vexil context
    │
    ▼
Phase 3: intent dispatcher
    │
    ├── Wardex: vulnerability triage sprint suggestion
    ├── Vexil:  entropy rule expansion suggestion
    └── NameSniper: refactoring session scope suggestion
```

nullcal is the **operational layer** of the portfolio. Wardex and Vexil are the security layer.  
Together they demonstrate: system design thinking, local-first architecture, and LLM orchestration without LLM dependency.
