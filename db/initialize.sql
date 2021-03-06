CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    location VARCHAR(255),
    description VARCHAR(255),
    url VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT FALSE
);