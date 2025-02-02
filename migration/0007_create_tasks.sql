CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    goal_id INT NOT NULL,
    phase_id INT,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'in-progress',
    estimated_time INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_task_goal
    FOREIGN KEY (goal_id)
    REFERENCES goals(id)
    ON DELETE CASCADE,

    CONSTRAINT fk_task_phase
    FOREIGN KEY (phase_id)
    REFERENCES phases(id)
    ON DELETE CASCADE
);
