ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45) NULL;

COMMENT ON COLUMN users.last_login_ip;
