DROP TABLE IF EXISTS audit_logs;

ALTER TABLE articles
    DROP COLUMN review_comment;
