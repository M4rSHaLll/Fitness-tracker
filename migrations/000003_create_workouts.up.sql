CREATE TABLE workouts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    description TEXT NOT NULL CHECK (length(trim(description)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX workouts_user_id_idx ON workouts (user_id);
