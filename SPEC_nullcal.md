# nullcal — Project Specification

> A TUI-native calendar and task manager. Google Calendar as a replaceable backend.  
> Your attention is the interface. Everything else is noise.

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
- Week view TUI navigable entirely by keyboard
- Local task management (create, edit, complete, delete)
- Fixed routine blocks (e.g. Monday = Vexil, Tuesday = Wardex)
- Google Calendar read/write via OAuth2
- Conflict detection between local tasks and GCal events

### Phase 2
- Full CRUD for GCal events from within the TUI
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
├── cmd/nullcal/
│   └── main.go                  ← entry point, signal handling, graceful shutdown
├── internal/
│   ├── config/                  ← env-based config, XDG paths, OAuth token storage
│   ├── store/                   ← SQLite repository (tasks, routine blocks, sync state)
│   ├── sync/
│   │   ├── adapter.go           ← CalendarAdapter interface
│   │   ├── gcal/                ← Google Calendar implementation
│   │   └── caldav/              ← CalDAV implementation (Phase 2)
│   ├── tui/
│   │   ├── model.go             ← Bubbletea root model
│   │   ├── weekview.go          ← 7-column pixel week grid
│   │   ├── taskpanel.go         ← side panel: today's tasks
│   │   ├── editor.go            ← modal: create/edit task or event
│   │   └── keys.go              ← keybindings registry
│   └── intent/                  ← Phase 3: Ollama dispatcher (stub in MVP)
├── pkg/
│   └── timeutil/                ← week boundaries, slot math, timezone helpers
├── Makefile
├── Dockerfile
├── .github/workflows/ci.yml
├── lazygo.yml
└── README.md
```

### Layer Responsibilities

| Layer | Owns | Does NOT own |
|---|---|---|
| `store` | All persistence, SQLite schema, migrations | Business logic, sync decisions |
| `sync/gcal` | GCal API calls, OAuth2 token lifecycle | Local state, conflict resolution |
| `tui` | All rendering, keybindings, modal flows | Direct DB access, API calls |
| `intent` | Ollama prompt construction, response parsing | Executing any calendar/task operations |

---

## 4. Core Types

```go
// Task is a local unit of work, independent of any external calendar.
type Task struct {
    ID          string
    Title       string
    Description string
    ProjectTag  string    // e.g. "wardex", "vexil", "namesniper"
    DueAt       time.Time
    CompletedAt *time.Time
    Recurrence  Recurrence
    CreatedAt   time.Time
}

// RoutineBlock is a fixed weekly time slot tied to a project.
// Defined once in config, rendered on every matching weekday.
type RoutineBlock struct {
    Weekday   time.Weekday
    StartTime string // "09:00"
    EndTime   string // "11:00"
    Label     string // "Wardex — feature dev"
    ProjectTag string
}

// CalendarEvent is the normalised representation of an external event.
// Populated by any CalendarAdapter implementation.
type CalendarEvent struct {
    ExternalID  string
    Source      string // "gcal", "caldav"
    Title       string
    StartAt     time.Time
    EndAt       time.Time
    Description string
    SyncedAt    time.Time
}

// CalendarAdapter is the interface every sync backend must satisfy.
// Defined here (consumer package), not in the adapter packages.
type CalendarAdapter interface {
    ListEvents(ctx context.Context, from, to time.Time) ([]CalendarEvent, error)
    CreateEvent(ctx context.Context, e CalendarEvent) (CalendarEvent, error)
    UpdateEvent(ctx context.Context, e CalendarEvent) error
    DeleteEvent(ctx context.Context, externalID string) error
}
```

---

## 5. TUI Design

Visual identity mirrors the @core image: pixel-font header, monochrome palette, high-density information layout with zero decorative chrome.

```
┌─────────────────────────────────────────────────────────────────────┐
│  nullcal                                    2026-03-09  week 11     │
├──────────┬──────────┬──────────┬──────────┬──────────┬──────────┬───┤
│  MON 09  │  TUE 10  │  WED 11  │  THU 12  │  FRI 13  │  SAT 14  │ SU│
├──────────┼──────────┼──────────┼──────────┼──────────┼──────────┼───┤
│ ░ Vexil  │ ░ Wardex │ ░ Name-  │ ░ Wardex │          │          │   │
│   09-11  │   09-11  │   Sniper │   09-11  │          │          │   │
│          │          │   09-11  │          │          │          │   │
│ · task 1 │          │          │ · task 3 │          │          │   │
│ ✓ task 2 │ · task 4 │          │          │          │          │   │
└──────────┴──────────┴──────────┴──────────┴──────────┴──────────┴───┘
 [n] new task  [e] edit  [d] delete  [s] sync  [q] quit  [?] help
```

**Keybindings (MVP)**

| Key | Action |
|---|---|
| `h` / `l` | Navigate week backward / forward |
| `j` / `k` | Move between tasks in focused column |
| `n` | New task (opens editor modal) |
| `e` | Edit selected task |
| `x` | Toggle task complete |
| `D` | Delete task (requires confirmation) |
| `s` | Trigger GCal sync |
| `q` | Quit |
| `?` | Help overlay |

---

## 6. Google Calendar OAuth2 Flow

OAuth2 tokens are stored at `$XDG_CONFIG_HOME/nullcal/token.json`.  
On first run, nullcal opens the browser for consent and writes the token.  
Subsequent runs refresh silently using the stored refresh token.

The token file **must** be in `.gitignore`. A `SECURITY.md` note will warn explicitly.

```
$XDG_CONFIG_HOME/nullcal/
├── config.yaml      ← user preferences, routine blocks, adapter selection
└── token.json       ← OAuth2 token (never committed)
```

---

## 7. SQLite Schema (MVP)

```sql
CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT,
    project_tag  TEXT,
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
```

WAL mode enabled on connection open. Foreign keys enforced.

---

## 8. Dependency Stack

| Dependency | Purpose | Rationale |
|---|---|---|
| `github.com/charmbracelet/bubbletea` | TUI framework | Already in Arcanum stack |
| `github.com/charmbracelet/lipgloss` | TUI styling | Same |
| `github.com/mattn/go-sqlite3` | SQLite driver | Proven, WAL-compatible |
| `google.golang.org/api/calendar/v3` | GCal API client | Official SDK |
| `golang.org/x/oauth2` | OAuth2 token management | stdlib-tier quality |
| `github.com/spf13/cobra` | CLI entrypoint | Consistent with lazy.go/Wardex |

No ORM. Raw SQL via `database/sql`. Queries are readable and auditable.

---

## 9. Phase 3 — Intent Dispatcher (Design Preview)

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
TUI renders proposal panel → user accepts / rejects each item
         │
         ▼
store.CreateTask() / gcal.CreateEvent()  ← only on explicit accept
```

The LLM never writes to the store or calls any adapter directly.  
This is the same invariant as Arcanum: **narrative only, no mechanical authority**.

---

## 10. Bootstrap with lazy.go

Save as `lazygo.yml` and run `lazy.go init --from lazygo.yml`:

```yaml
# nullcal — lazygo.yml
# Bootstrap: lazy.go init --from lazygo.yml

project:
  name: nullcal
  module_path: github.com/had-nu/nullcal
  description: "TUI-native calendar and task manager. Google Calendar as a replaceable backend."
  author: had-nu
  type: cli
  license: apache-2.0
  visibility: public
  criticality: experimental

features:
  docker: false
  github_actions: true
  linting: true
  static_analysis: true
  sast: false
  dependabot: true
  tests: true

github:
  enabled: true
  topics:
    - go
    - tui
    - calendar
    - bubbletea
    - productivity
    - local-first
  push_on_init: false
```

> `push_on_init: false` — revisa a estrutura gerada antes de subir.  
> `sast: false` — nullcal não processa inputs adversariais; gosec manual é suficiente.  
> `docker: false` — ferramenta local, não há runtime containerizado.

---

## 11. Development Phases & Milestones

| Phase | Milestone | Est. Effort |
|---|---|---|
| 0 | Scaffold via lazy.go, go.mod, SQLite connection, config loader | 1 session |
| 1a | Store layer: CRUD for tasks and routine blocks | 1–2 sessions |
| 1b | TUI: static week view with hardcoded data | 1 session |
| 1c | TUI: week view wired to store | 1 session |
| 1d | GCal adapter: OAuth2 flow + ListEvents | 2 sessions |
| 1e | Sync: write CalendarEvents to local store, render in week view | 1 session |
| 1f | GCal write: CreateEvent from TUI editor | 1 session |
| 2a | CalDAV adapter (same interface as gcal) | 2 sessions |
| 2b | Notification layer | 1 session |
| 3 | Intent dispatcher + Ollama integration | TBD |

---

## 12. Relation to the Ecosystem

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

---

*Spec version: 1.0 — March 2026*  
*Next review: after Phase 1e (GCal sync functional)*
