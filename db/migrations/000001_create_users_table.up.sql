CREATE TABLE users
(
    id                SERIAL PRIMARY KEY,
    username          VARCHAR(50)              NOT NULL UNIQUE,
    first_name        varchar(50),
    last_name         varchar(80),
    phone             VARCHAR(20)              NOT NULL UNIQUE,
    password          VARCHAR(255)             NOT NULL,
    last_seen         TIMESTAMP WITH TIME ZONE,
    profile_picture   varchar(50),
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    phone_verified_at TIMESTAMP WITH TIME ZONE
);