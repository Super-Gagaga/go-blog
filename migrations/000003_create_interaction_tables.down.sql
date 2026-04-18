DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS comments;

ALTER TABLE articles
    DROP COLUMN favorite_count,
    DROP COLUMN like_count,
    DROP COLUMN comment_count;
