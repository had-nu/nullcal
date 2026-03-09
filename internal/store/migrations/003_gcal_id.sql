-- Add gcal_event_id to tasks to link them with pushed Google Calendar events.
-- This allows updating and deleting the external event when the task changes.
ALTER TABLE tasks ADD COLUMN gcal_event_id TEXT;
