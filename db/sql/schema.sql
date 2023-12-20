-- SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
--
-- SPDX-License-Identifier: CC0-1.0

CREATE TABLE users
(
    username   TEXT      NOT NULL PRIMARY KEY,
    hash       TEXT      NOT NULL,
    salt       TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE sessions
(
    token      TEXT      NOT NULL PRIMARY KEY,
    username   TEXT      NOT NULL,
    expires    TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE projects
(
    url        TEXT      NOT NULL PRIMARY KEY,
    name       TEXT      NOT NULL,
    forge      TEXT      NOT NULL,
    version    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE releases
(
    id          TEXT      NOT NULL PRIMARY KEY,
    project_url TEXT      NOT NULL,
    release_url TEXT      NOT NULL,
    tag         TEXT      NOT NULL,
    content     TEXT      NOT NULL,
    date        TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
