CREATE SCHEMA todoapp;

CREATE TABLE todoapp.users
(
    id           INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    version      BIGINT       NOT NULL DEFAULT 1,
    full_name    VARCHAR(100) NOT NULL,
    phone_number VARCHAR(15),

    CONSTRAINT full_name_min_length CHECK (char_length(full_name) >= 3),
    CONSTRAINT phone_number_format CHECK (phone_number ~ '^\+[0-9]+$'
) ,
    CONSTRAINT phone_number_length   CHECK (char_length(phone_number) BETWEEN 10 AND 15)
);

CREATE TABLE todoapp.tasks
(
    id             INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    version        BIGINT       NOT NULL DEFAULT 1,
    title          VARCHAR(100) NOT NULL,
    description    VARCHAR(1000),
    completed      BOOLEAN      NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL,
    completed_at   TIMESTAMPTZ,

    author_user_id INTEGER      NOT NULL REFERENCES todoapp.users (id),

    CONSTRAINT title_length CHECK ( char_length(title) BETWEEN 1 AND 100),
    CONSTRAINT description_length CHECK ( char_length(description) BETWEEN 1 AND 1000),

    CHECK (
        (completed = FALSE AND completed_at IS NULL)
            OR
        (completed = TRUE AND completed_at IS NOT NULL AND completed_at >= created_at)
        )
);