CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    login      VARCHAR(255) NOT NULL,
    pass       VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_login_idx ON users (login);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE set_updated_at();

CREATE OR REPLACE FUNCTION set_users_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.created_at IS NULL THEN
        NEW.created_at = CURRENT_TIMESTAMP;
    END IF;
    IF NEW.updated_at IS NULL THEN
        NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_created_at
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE PROCEDURE set_users_timestamps();
