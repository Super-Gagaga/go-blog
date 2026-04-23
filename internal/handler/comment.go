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

type CommentHandler struct {
	comments *service.CommentService
}

func NewCommentHandler(comments *service.CommentService) *CommentHandler {
	return &CommentHandler{comments: comments}
}

func (h *CommentHandler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req service.CreateCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	comment, err := h.comments.CreateComment(c.Request.Context(), userID, articleID, req)
	if err != nil {
		writeCommentError(c, err)
		return
	}
	app.Success(c, comment)
}

func (h *CommentHandler) Reply(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	commentID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req service.ReplyCommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	comment, err := h.comments.ReplyComment(c.Request.Context(), userID, commentID, req)
	if err != nil {
		writeCommentError(c, err)
		return
	}
	app.Success(c, comment)
}

func (h *CommentHandler) Delete(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	commentID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.comments.DeleteComment(c.Request.Context(), userID, commentID); err != nil {
		writeCommentError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *CommentHandler) ListByArticle(c *gin.Context) {
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	comments, err := h.comments.GetArticleComments(c.Request.Context(), articleID)
	if err != nil {
		writeCommentError(c, err)
		return
	}
	app.Success(c, comments)
}

func writeCommentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrArticleNotFound), errors.Is(err, service.ErrCommentNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	case errors.Is(err, service.ErrCommentForbidden):
		app.Error(c, http.StatusForbidden, errcode.Forbidden, err.Error())
	case errors.Is(err, service.ErrInvalidComment):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
