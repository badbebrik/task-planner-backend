CREATE TABLE IF NOT EXISTS phases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'in-progress',
    estimated_time INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_phase_goal
    FOREIGN KEY (goal_id)
    REFERENCES goals(id)
    ON DELETE CASCADE
);
