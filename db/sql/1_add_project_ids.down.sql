-- SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
--
-- SPDX-License-Identifier: CC0-1.0

--ALTER TABLE projects RENAME TO projects_tmp; -- noqa

ALTER TABLE projects RENAME TO projects_tmp;

CREATE TABLE IF NOT EXISTS projects (
    url TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    forge TEXT NOT NULL,
    version TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO projects (url, name, forge, version, created_at)
SELECT
    url,
    name,
    forge,
    version,
    created_at
FROM projects_tmp;

DROP TABLE projects_tmp;
