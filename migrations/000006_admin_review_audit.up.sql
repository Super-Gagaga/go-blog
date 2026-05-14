ALTER TABLE articles
    ADD COLUMN review_comment VARCHAR(500) NULL AFTER status;

CREATE TABLE audit_logs (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    admin_id BIGINT NOT NULL,
    action VARCHAR(80) NOT NULL,
    target_type VARCHAR(40),
    target_id BIGINT,
    detail TEXT,
    ip VARCHAR(64),
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_audit_logs_admin_id (admin_id),
    KEY idx_audit_logs_action (action),
    KEY idx_audit_logs_target_type (target_type),
    KEY idx_audit_logs_target_id (target_id),
    KEY idx_audit_logs_created_at (created_at),
    CONSTRAINT fk_audit_logs_admin FOREIGN KEY (admin_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
