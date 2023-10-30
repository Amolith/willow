-- SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
--
-- SPDX-License-Identifier: CC0-1.0

-- Create table of users with username, password hash, salt, and creation
-- timestamp
CREATE TABLE users
(
    username   TEXT      NOT NULL PRIMARY KEY,
    hash       TEXT      NOT NULL,
    salt       TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create table of sessions with session GUID, username, and timestamp of when
-- the session was created
CREATE TABLE sessions
(
    token      TEXT      NOT NULL PRIMARY KEY,
    username   TEXT      NOT NULL,
    expires    TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create table of tracked projects with URL, name, forge, running version, and
-- timestamp of when the project was added
CREATE TABLE projects
(
    url        TEXT      NOT NULL PRIMARY KEY,
    name       TEXT      NOT NULL,
    forge      TEXT      NOT NULL,
    version    TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create table of project releases with the project URL and the release tags,
-- contents, URLs, and dates
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
