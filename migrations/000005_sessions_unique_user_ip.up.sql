ALTER TABLE sessions
    ADD CONSTRAINT sessions_user_id_ip_key UNIQUE (user_id, ip);
