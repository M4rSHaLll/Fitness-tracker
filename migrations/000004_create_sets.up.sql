CREATE TABLE sets (
    id BIGSERIAL PRIMARY KEY,
    exercise_id BIGINT NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    workout_id BIGINT NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    reps BIGINT NOT NULL CHECK (reps >= 0 AND reps <= 1000),
    weight DOUBLE PRECISION NOT NULL CHECK (weight >= 1 AND weight <= 10000),
    rpe NUMERIC(3, 1) NOT NULL CHECK (rpe >= 6 AND rpe <= 10 AND rpe * 2 = floor(rpe * 2)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sets_workout_id_idx ON sets (workout_id);
CREATE INDEX sets_exercise_id_idx ON sets (exercise_id);
