CREATE TABLE categories (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_categories_slug (slug)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE tags (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    UNIQUE KEY uk_tags_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE articles (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(300) NOT NULL,
    content TEXT NOT NULL,
    content_html TEXT NOT NULL,
    summary VARCHAR(500),
    cover_image VARCHAR(500),
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    view_count BIGINT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    category_id BIGINT,
    created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    deleted_at DATETIME(3),
    UNIQUE KEY uk_articles_slug (slug),
    KEY idx_articles_user_id (user_id),
    KEY idx_articles_status (status),
    KEY idx_articles_category_id (category_id),
    KEY idx_articles_created_at (created_at),
    KEY idx_articles_deleted_at (deleted_at),
    FULLTEXT KEY ft_articles_search (title, content),
    CONSTRAINT fk_articles_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_articles_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE article_tags (
    article_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL,
    PRIMARY KEY (article_id, tag_id),
    CONSTRAINT fk_article_tags_article FOREIGN KEY (article_id) REFERENCES articles(id) ON DELETE CASCADE,
    CONSTRAINT fk_article_tags_tag FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO categories (name, slug) VALUES
    ('Go', 'go'),
    ('Backend', 'backend'),
    ('Architecture', 'architecture');

INSERT INTO tags (name) VALUES
    ('go'),
    ('gin'),
    ('gorm'),
    ('mysql'),
    ('redis');
