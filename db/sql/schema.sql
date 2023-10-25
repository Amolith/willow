-- SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
--
-- SPDX-License-Identifier: CC0-1.0

-- Create table of users with username, password hash, salt, and creation
-- timestamp
CREATE TABLE users (
    username       VARCHAR(255)  NOT         NULL,
    hash           VARCHAR(255)  NOT         NULL,
    salt           VARCHAR(255)  NOT         NULL,
    created_at     TIMESTAMP     NOT         NULL   DEFAULT  CURRENT_TIMESTAMP,
    PRIMARY        KEY           (username)
);

-- Create table of sessions with session GUID, username, and timestamp of when
-- the session was created
CREATE TABLE sessions (
    token        VARCHAR(255)  NOT          NULL,
    username     VARCHAR(255)  NOT          NULL,
    expires      TIMESTAMP     NOT          NULL,
    created_at   TIMESTAMP     NOT          NULL   DEFAULT  CURRENT_TIMESTAMP,
    PRIMARY      KEY           (token)
);

-- Create table of tracked projects with URL, name, forge, running version, and
-- timestamp of when the project was added
CREATE TABLE projects (
    url         VARCHAR(255)  NOT    NULL,
    name        VARCHAR(255)  NOT    NULL,
    forge       VARCHAR(255)  NOT    NULL,
    version     VARCHAR(255)  NOT    NULL,
    created_at  TIMESTAMP     NOT    NULL   DEFAULT  CURRENT_TIMESTAMP,
    PRIMARY     KEY           (url)
);

-- Create table of project releases with the project URL and the release tags,
-- contents, URLs, and dates
CREATE TABLE releases (
    project_url  VARCHAR(255)  NOT    NULL,
    release_url  VARCHAR(255)  NOT    NULL,
    tag          VARCHAR(255)  NOT    NULL,
    content      TEXT          NOT    NULL,
    date         TIMESTAMP     NOT    NULL,
    created_at   TIMESTAMP     NOT    NULL   DEFAULT  CURRENT_TIMESTAMP,
    PRIMARY      KEY           (release_url)
);