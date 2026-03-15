CREATE TABLE IF NOT EXISTS withdrawals (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    order_id   VARCHAR(48)   NOT NULL,
    sum        NUMERIC(20, 2) NOT NULL,
    created_at TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX withdrawals_user_id_idx ON withdrawals (user_id);

CREATE TRIGGER withdrawals_updated_at
    BEFORE UPDATE ON withdrawals
    FOR EACH ROW
    EXECUTE PROCEDURE set_updated_at();
