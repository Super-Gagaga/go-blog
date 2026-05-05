ALTER TABLE users
    ADD COLUMN article_count BIGINT NOT NULL DEFAULT 0 AFTER role,
    ADD COLUMN follower_count BIGINT NOT NULL DEFAULT 0 AFTER article_count,
    ADD COLUMN following_count BIGINT NOT NULL DEFAULT 0 AFTER follower_count;

UPDATE users
SET article_count = (
    SELECT COUNT(*)
    FROM articles
    WHERE articles.user_id = users.id
      AND articles.status = 'published'
      AND articles.deleted_at IS NULL
);

CREATE TABLE follows (
    follower_id BIGINT NOT NULL,
    followee_id BIGINT NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (follower_id, followee_id),
    KEY idx_follows_followee_id (followee_id),
    CONSTRAINT fk_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_follows_followee FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE notifications (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(30) NOT NULL,
    content JSON NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    KEY idx_notifications_user_read (user_id, is_read),
    KEY idx_notifications_created_at (created_at),
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
