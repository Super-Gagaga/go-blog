package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/middleware"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type FollowHandler struct {
	follows *service.FollowService
}

func NewFollowHandler(follows *service.FollowService) *FollowHandler {
	return &FollowHandler{follows: follows}
}

func (h *FollowHandler) Toggle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	followeeID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	following, err := h.follows.ToggleFollow(c.Request.Context(), userID, followeeID)
	if err != nil {
		writeFollowError(c, err)
		return
	}
	app.Success(c, gin.H{"following": following})
}

func (h *FollowHandler) Following(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	result, err := h.follows.GetFollowing(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "page_size", 10))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, result)
}

func (h *FollowHandler) Followers(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	result, err := h.follows.GetFollowers(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "page_size", 10))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, result)
}

func writeFollowError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCannotFollowSelf):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrFollowUserNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
