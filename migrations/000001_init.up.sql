CREATE TABLE IF NOT EXISTS clients (
    id TEXT PRIMARY KEY,
    capacity INTEGER NOT NULL,
    refill_rate INTEGER NOT NULL
);