CREATE TABLE IF NOT EXISTS files (
    id SERIAL PRIMARY KEY,
    student_id UUID NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL
);
