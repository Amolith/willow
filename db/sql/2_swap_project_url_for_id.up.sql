-- SPDX-FileCopyrightText: Amolith <amolith@secluded.site>
--
-- SPDX-License-Identifier: CC0-1.0

ALTER TABLE releases RENAME TO releases_tmp;

CREATE TABLE IF NOT EXISTS releases
(
    id         TEXT      NOT NULL PRIMARY KEY,
    project_id TEXT      NOT NULL,
    url        TEXT      NOT NULL,
    tag        TEXT      NOT NULL,
    content    TEXT      NOT NULL,
    date       TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO releases (id, project_id, url, tag, content, date)
SELECT
    r.id,
    p.id,
    r.release_url,
    r.tag,
    r.content,
    r.date
FROM releases_tmp r
JOIN projects p ON r.project_url = p.url;

DROP TABLE releases_tmp;