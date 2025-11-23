CREATE TABLE IF NOT EXISTS users
(
    id SERIAL PRIMARY KEY,
    username VARCHAR(25) NOT NULL,
    team_name VARCHAR(25) REFERENCES teams (name),
    is_active BOOLEAN NOT NULL
);