CREATE TABLE users (
       id SERIAL PRIMARY KEY,
       email VARCHAR(255) UNIQUE NOT NULL,
       password_hash TEXT NOT NULL,
       name VARCHAR(255),
       is_email_verified BOOLEAN DEFAULT FALSE,
       google_id VARCHAR(255),
       created_at TIMESTAMP DEFAULT NOW(),
       updated_at TIMESTAMP DEFAULT NOW()
);
