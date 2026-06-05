package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type RecommendationHandler struct {
	recommendations *service.RecommendationService
}

func NewRecommendationHandler(recommendations *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recommendations: recommendations}
}

// RecommendedAuthors returns a list of recommended authors.
func (h *RecommendationHandler) RecommendedAuthors(c *gin.Context) {
	authors, err := h.recommendations.GetRecommendedAuthors(c.Request.Context(), queryInt(c, "limit", 5))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, authors)
}

// TrendingTopics returns a list of trending topics (tags with hotness scores).
func (h *RecommendationHandler) TrendingTopics(c *gin.Context) {
	topics, err := h.recommendations.GetTrendingTags(c.Request.Context(), queryInt(c, "limit", 12))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, topics)
}
