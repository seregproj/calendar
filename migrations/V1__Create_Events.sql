CREATE TABLE events (
    id uuid,
    title VARCHAR NOT NULL,
    description VARCHAR NOT NULL,
    datetime_start TIMESTAMP,
    datetime_finish TIMESTAMP,
    processed BOOLEAN NOT NULL DEFAULT false,
    date_add TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
);
