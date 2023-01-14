CREATE TABLE IF NOT EXISTS mtpp_table (
    id SERIAL PRIMARY KEY NOT NULL,
    mtpp_number VARCHAR NOT NULL,
    mtpp_url VARCHAR NOT NULL,
    version int NOT NULL DEFAULT 1,
    created_at timestamptz DEFAULT current_timestamp
);