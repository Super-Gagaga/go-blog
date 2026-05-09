package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/middleware"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type FeedHandler struct {
	feeds *service.FeedService
}

func NewFeedHandler(feeds *service.FeedService) *FeedHandler {
	return &FeedHandler{feeds: feeds}
}

func (h *FeedHandler) UserFeed(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	result, err := h.feeds.GetUserFeed(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "page_size", 10))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, result)
}
