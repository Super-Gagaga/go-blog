package cache

import "fmt"

func ArticleLikeSetKey(articleID int64) string {
	return fmt.Sprintf("article:%d:likes", articleID)
}

func ArticleFavoriteSetKey(articleID int64) string {
	return fmt.Sprintf("article:%d:favorites", articleID)
}

func ArticleCommentCountKey(articleID int64) string {
	return fmt.Sprintf("article:%d:comment_count", articleID)
}

// ArticleCacheKey 文章详情的缓存旁路键(cachemanager.ServiceManager[model.Article])
func ArticleCacheKey(articleID int64) string {
	return fmt.Sprintf("article:%d", articleID)
}

func ArticleListKey(signature string) string {
	return fmt.Sprintf("articles:list:%s", signature)
}

// ArticleListKeyPattern 匹配全部列表缓存键,用于写后按模式批量失效
const ArticleListKeyPattern = "articles:list:*"

// UserCacheKey 用户资料的缓存旁路键(cachemanager.ServiceManager[model.User])
func UserCacheKey(userID int64) string {
	return fmt.Sprintf("user:%d", userID)
}

func HotFeedKey() string {
	return "feed:hot"
}

func UserFollowingSetKey(userID int64) string {
	return fmt.Sprintf("user:%d:following", userID)
}

func UserFollowersSetKey(userID int64) string {
	return fmt.Sprintf("user:%d:followers", userID)
}

func UserUnreadNotificationsKey(userID int64) string {
	return fmt.Sprintf("user:%d:notifications:unread", userID)
}

func UserFeedKey(userID int64) string {
	return fmt.Sprintf("user:%d:feed", userID)
}
