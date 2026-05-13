package model

type AuthorStat struct {
	UserID       int64  `json:"user_id"`
	Username     string `json:"username"`
	Avatar       string `json:"avatar"`
	ArticleCount int64  `json:"article_count"`
}

type DashboardStats struct {
	UserCount          int64        `json:"user_count"`
	AdminCount         int64        `json:"admin_count"`
	ArticleCount       int64        `json:"article_count"`
	PublishedCount     int64        `json:"published_count"`
	PendingReviewCount int64        `json:"pending_review_count"`
	DraftCount         int64        `json:"draft_count"`
	CommentCount       int64        `json:"comment_count"`
	LikeCount          int64        `json:"like_count"`
	FavoriteCount      int64        `json:"favorite_count"`
	FollowCount        int64        `json:"follow_count"`
	RecentUsers        []User       `json:"recent_users"`
	RecentArticles     []Article    `json:"recent_articles"`
	TopAuthors         []AuthorStat `json:"top_authors"`
}
