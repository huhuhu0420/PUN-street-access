CREATE TABLE
    IF NOT EXISTS categorys (
        category_id SERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL
    );