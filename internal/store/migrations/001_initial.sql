-- Initial schema for nullcal.

CREATE TABLE tasks (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT,
    project_tag  TEXT,
    status       TEXT NOT NULL DEFAULT 'backlog',
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
    cursor       TEXT  -- adapter-defined opaque continuation token
);
