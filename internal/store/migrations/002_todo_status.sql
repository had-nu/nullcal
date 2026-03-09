-- Add 'todo' as a pre-backlog status.
-- Tasks in 'todo' appear in the To-Do List pane but NOT in the Kanban board.
-- Tasks in 'backlog' appear in the Kanban Backlog column but NOT in the To-Do List.

-- Migrate existing tasks: backlog → todo (they should still show in todo pane).
-- This preserves all existing data — only affect how the UI filters them.
-- No structural change needed; SQLite TEXT column already accepts any value.
UPDATE tasks SET status = 'todo' WHERE status = 'backlog';
