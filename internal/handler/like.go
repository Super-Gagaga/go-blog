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

type LikeHandler struct {
	likes *service.LikeService
}

func NewLikeHandler(likes *service.LikeService) *LikeHandler {
	return &LikeHandler{likes: likes}
}

func (h *LikeHandler) Toggle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	liked, count, err := h.likes.ToggleLike(c.Request.Context(), userID, articleID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}

	app.Success(c, gin.H{
		"liked": liked,
		"count": count,
	})
}

func writeInteractionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInteractionTargetNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
