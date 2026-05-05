DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS follows;

ALTER TABLE users
    DROP COLUMN following_count,
    DROP COLUMN follower_count,
    DROP COLUMN article_count;
