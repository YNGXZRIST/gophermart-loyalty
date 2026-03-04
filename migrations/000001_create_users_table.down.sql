DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TRIGGER IF EXISTS users_created_at ON users;

DROP FUNCTION IF EXISTS set_updated_at();
DROP FUNCTION IF EXISTS set_users_timestamps();

DROP INDEX IF EXISTS users_login_idx;
DROP TABLE IF EXISTS users;
