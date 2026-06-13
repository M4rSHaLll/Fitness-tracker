CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT NOT NULL CHECK (length(trim(username)) > 0),
    weight DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (weight = 0 OR (weight >= 1 AND weight <= 500)),
    height DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (height = 0 OR (height >= 1 AND height <= 300)),
    age BIGINT NOT NULL DEFAULT 0 CHECK (age = 0 OR (age >= 1 AND age <= 150)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
