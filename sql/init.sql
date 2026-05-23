-- ============================================================================
-- go-blog Database Initialization Script
--
-- Target: MySQL 8.4+
-- Usage:  mysql -u root -p < sql/init.sql
--
-- This file consolidates all migrations into a single init script:
--   000001_create_users_table
--   000002_create_content_tables
--   000003_create_interaction_tables
--   000004_create_community_tables
--   000005_phase6_optimization
--   000006_admin_review_audit
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. users — 用户表
-- ----------------------------------------------------------------------------
CREATE TABLE users (
    id              BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    username        VARCHAR(50)  NOT NULL,
    email           VARCHAR(255) NOT NULL,
    password        VARCHAR(255) NOT NULL,
    avatar          VARCHAR(500)          DEFAULT NULL,
    bio             TEXT                  DEFAULT NULL,
    role            VARCHAR(20)  NOT NULL DEFAULT 'user',
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    article_count   BIGINT       NOT NULL DEFAULT 0,
    follower_count  BIGINT       NOT NULL DEFAULT 0,
    following_count BIGINT       NOT NULL DEFAULT 0,
    created_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at      DATETIME(3)          DEFAULT NULL,

    UNIQUE KEY uk_users_username (username),
    UNIQUE KEY uk_users_email    (email),
    KEY       idx_users_deleted_at (deleted_at),
    KEY       idx_users_status     (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 2. categories — 文章分类
-- ----------------------------------------------------------------------------
CREATE TABLE categories (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    slug       VARCHAR(100) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 3. tags — 文章标签
-- ----------------------------------------------------------------------------
CREATE TABLE tags (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(50) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    UNIQUE KEY uk_tags_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 4. articles — 文章
-- ----------------------------------------------------------------------------
CREATE TABLE articles (
    id             BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title          VARCHAR(255) NOT NULL,
    slug           VARCHAR(300) NOT NULL,
    content        TEXT         NOT NULL,
    content_html   TEXT         NOT NULL,
    summary        VARCHAR(500)          DEFAULT NULL,
    cover_image    VARCHAR(500)          DEFAULT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'draft',
    review_comment VARCHAR(500)          DEFAULT NULL,
    view_count     BIGINT       NOT NULL DEFAULT 0,
    comment_count  BIGINT       NOT NULL DEFAULT 0,
    like_count     BIGINT       NOT NULL DEFAULT 0,
    favorite_count BIGINT       NOT NULL DEFAULT 0,
    user_id        BIGINT       NOT NULL,
    category_id    BIGINT                DEFAULT NULL,
    created_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at     DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at     DATETIME(3)           DEFAULT NULL,

    UNIQUE KEY uk_articles_slug        (slug),
    KEY       idx_articles_user_id     (user_id),
    KEY       idx_articles_status      (status),
    KEY       idx_articles_category_id (category_id),
    KEY       idx_articles_created_at  (created_at),
    KEY       idx_articles_deleted_at  (deleted_at),
    FULLTEXT  KEY ft_articles_search   (title, content),

    CONSTRAINT fk_articles_user     FOREIGN KEY (user_id)     REFERENCES users(id),
    CONSTRAINT fk_articles_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 5. article_tags — 文章⇔标签关联表 (M:N)
-- ----------------------------------------------------------------------------
CREATE TABLE article_tags (
    article_id BIGINT NOT NULL,
    tag_id     BIGINT NOT NULL,

    PRIMARY KEY (article_id, tag_id),

    CONSTRAINT fk_article_tags_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_tags_tag     FOREIGN KEY (tag_id)     REFERENCES tags(id)     ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 6. comments — 评论 (两级嵌套)
-- ----------------------------------------------------------------------------
CREATE TABLE comments (
    id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
    content    TEXT         NOT NULL,
    article_id BIGINT       NOT NULL,
    user_id    BIGINT       NOT NULL,
    parent_id  BIGINT                DEFAULT NULL,
    created_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3)           DEFAULT NULL,

    KEY idx_comments_article_id (article_id),
    KEY idx_comments_user_id    (user_id),
    KEY idx_comments_parent_id  (parent_id),
    KEY idx_comments_deleted_at (deleted_at),

    CONSTRAINT fk_comments_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT fk_comments_user    FOREIGN KEY (user_id)    REFERENCES users(id),
    CONSTRAINT fk_comments_parent  FOREIGN KEY (parent_id)  REFERENCES comments(id)  ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 7. likes — 点赞 (userId + articleId 联合主键)
-- ----------------------------------------------------------------------------
CREATE TABLE likes (
    user_id    BIGINT      NOT NULL,
    article_id BIGINT      NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (user_id, article_id),
    KEY idx_likes_article_id (article_id),

    CONSTRAINT fk_likes_user    FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE,
    CONSTRAINT fk_likes_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 8. favorites — 收藏 (userId + articleId 联合主键)
-- ----------------------------------------------------------------------------
CREATE TABLE favorites (
    user_id    BIGINT      NOT NULL,
    article_id BIGINT      NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (user_id, article_id),
    KEY idx_favorites_article_id     (article_id),
    KEY idx_favorites_user_created_at (user_id, created_at),

    CONSTRAINT fk_favorites_user    FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE,
    CONSTRAINT fk_favorites_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 9. follows — 关注关系 (followerId + followeeId 联合主键)
-- ----------------------------------------------------------------------------
CREATE TABLE follows (
    follower_id BIGINT      NOT NULL,
    followee_id BIGINT      NOT NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (follower_id, followee_id),
    KEY idx_follows_followee_id (followee_id),

    CONSTRAINT fk_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_follows_followee FOREIGN KEY (followee_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 10. notifications — 通知
-- ----------------------------------------------------------------------------
CREATE TABLE notifications (
    id         BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT      NOT NULL,
    type       VARCHAR(30) NOT NULL,
    content    JSON        NOT NULL,
    is_read    BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    KEY idx_notifications_user_read  (user_id, is_read),
    KEY idx_notifications_created_at (created_at),

    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------------------------------------------------------
-- 11. audit_logs — 管理员操作审计日志
-- ----------------------------------------------------------------------------
CREATE TABLE audit_logs (
    id          BIGINT      NOT NULL AUTO_INCREMENT PRIMARY KEY,
    admin_id    BIGINT      NOT NULL,
    action      VARCHAR(80) NOT NULL,
    target_type VARCHAR(40)          DEFAULT NULL,
    target_id   BIGINT               DEFAULT NULL,
    detail      TEXT                 DEFAULT NULL,
    ip          VARCHAR(64)          DEFAULT NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    KEY idx_audit_logs_admin_id    (admin_id),
    KEY idx_audit_logs_action      (action),
    KEY idx_audit_logs_target_type (target_type),
    KEY idx_audit_logs_target_id   (target_id),
    KEY idx_audit_logs_created_at  (created_at),

    CONSTRAINT fk_audit_logs_admin FOREIGN KEY (admin_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- Seed Data — 初始化分类和标签
-- ============================================================================
INSERT INTO categories (name, slug) VALUES
    ('Go',           'go'),
    ('Backend',      'backend'),
    ('Architecture', 'architecture');

INSERT INTO tags (name) VALUES
    ('go'),
    ('gin'),
    ('gorm'),
    ('mysql'),
    ('redis');
