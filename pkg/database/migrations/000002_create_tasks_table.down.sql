-- Drop trigger
DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;

-- Drop indexes
DROP INDEX IF EXISTS idx_tasks_user_id;

-- Drop tasks table
DROP TABLE IF EXISTS tasks;

-- Drop enum type
DROP TYPE IF EXISTS task_status;