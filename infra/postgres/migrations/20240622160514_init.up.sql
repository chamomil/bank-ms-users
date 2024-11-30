CREATE TABLE users
(
    id            BIGSERIAL PRIMARY KEY,
    uuid          UUID UNIQUE        NOT NULL DEFAULT gen_random_uuid(),
    login         VARCHAR(32) UNIQUE NOT NULL CHECK ( login ~ '^[a-z0-9_-]+$'),
    email         VARCHAR(32) UNIQUE NOT NULL CHECK ( email ~ '^.+@.+\..+$' ),
    password      BYTEA              NOT NULL CHECK ( length(password) <= 60 ),
    "telegramId"  BIGINT UNIQUE,
    "isActivated" BOOLEAN            NOT NULL DEFAULT FALSE,
    "createdAt"   TIMESTAMP          NOT NULL DEFAULT current_timestamp
);

CREATE TABLE users_auth_history
(
    id        BIGSERIAL PRIMARY KEY,
    "userId"  BIGINT       NOT NULL REFERENCES users (id),
    "agent"   VARCHAR(255) NOT NULL,
    ip        INET         NOT NULL,
    timestamp TIMESTAMP    NOT NULL DEFAULT current_timestamp
);

CREATE TABLE countries
(
    id   SERIAL PRIMARY KEY,
    code CHAR(2) UNIQUE      NOT NULL,
    name VARCHAR(128) UNIQUE NOT NULL
);

CREATE TABLE users_personal_data
(
    id              BIGINT PRIMARY KEY REFERENCES users (id),
    "phoneNumber"   CHAR(16)                               NOT NULL CHECK ("phoneNumber" ~ '^+[1-9][0-9]{1,14}$'),
    "firstName"     VARCHAR(64)                            NOT NULL,
    "lastName"      VARCHAR(64)                            NOT NULL,
    "fathersName"   VARCHAR(64),
    "dateOfBirth"   DATE                                   NOT NULL,
    "passportId"    VARCHAR(128)                           NOT NULL UNIQUE,
    "address"       VARCHAR(128)                           NOT NULL,
    "gender"        CHAR(1) CHECK ("gender" IN ('M', 'F')) NOT NULL,
    "liveInCountry" INTEGER                                NOT NULL REFERENCES countries (id)
);

CREATE TABLE workplaces
(
    "id"      BIGSERIAL PRIMARY KEY,
    "name"    VARCHAR(128) UNIQUE NOT NULL,
    "address" VARCHAR(255)        NOT NULL
);

CREATE TABLE users_employments
(
    "userId"      BIGINT       NOT NULL REFERENCES users (id),
    "workplaceId" BIGINT       NOT NULL REFERENCES workplaces (id),
    "position"    VARCHAR(128) NOT NULL,
    "startDate"   DATE         NOT NULL,
    "endDate"     DATE,
    PRIMARY KEY ("userId", "workplaceId")
);
