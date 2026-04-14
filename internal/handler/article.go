package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/middleware"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type ArticleHandler struct {
	articles *service.ArticleService
}

func NewArticleHandler(articles *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{articles: articles}
}

func (h *ArticleHandler) Create(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	var req service.CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	article, err := h.articles.CreateArticle(c.Request.Context(), userID, req)
	if err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, article)
}

func (h *ArticleHandler) Update(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req service.UpdateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	article, err := h.articles.UpdateArticle(c.Request.Context(), userID, articleID, req)
	if err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, article)
}

func (h *ArticleHandler) Delete(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.articles.DeleteArticle(c.Request.Context(), userID, articleID); err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, nil)
}

func (h *ArticleHandler) Get(c *gin.Context) {
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	article, err := h.articles.GetArticle(c.Request.Context(), articleID)
	if err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, article)
}

func (h *ArticleHandler) List(c *gin.Context) {
	req := service.ListArticleReq{
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 10),
		Status:     c.Query("status"),
		CategoryID: queryInt64(c, "category_id", 0),
		TagID:      queryInt64(c, "tag_id", 0),
		Keyword:    c.Query("keyword"),
		UserID:     queryInt64(c, "user_id", 0),
	}

	result, err := h.articles.ListArticles(c.Request.Context(), req)
	if err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, result)
}

func (h *ArticleHandler) MyArticles(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	req := service.ListArticleReq{
		Page:       queryInt(c, "page", 1),
		PageSize:   queryInt(c, "page_size", 10),
		Status:     c.Query("status"),
		CategoryID: queryInt64(c, "category_id", 0),
		TagID:      queryInt64(c, "tag_id", 0),
		Keyword:    c.Query("keyword"),
	}

	result, err := h.articles.ListUserArticles(c.Request.Context(), userID, req)
	if err != nil {
		writeArticleError(c, err)
		return
	}
	app.Success(c, result)
}

func (h *ArticleHandler) ChangeStatus(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req service.ChangeArticleStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	if err := h.articles.ChangeStatus(c.Request.Context(), userID, articleID, req.Status); err != nil {
		writeArticleError(c, err)
		return
	}

	app.Success(c, nil)
}

func (h *ArticleHandler) Ranking(c *gin.Context) {
	articles, err := h.articles.Ranking(c.Request.Context(), c.Query("period"), queryInt(c, "limit", 10))
	if err != nil {
		writeArticleError(c, err)
		return
	}
	app.Success(c, articles)
}

func (h *ArticleHandler) UploadImage(c *gin.Context) {
	if _, ok := middleware.CurrentUserID(c); !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "image is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "open image failed")
		return
	}
	defer file.Close()

	url, err := h.articles.UploadArticleImage(c.Request.Context(), file, fileHeader.Filename, fileHeader.Size)
	if err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	app.Success(c, gin.H{"url": url})
}

func writeArticleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrArticleNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	case errors.Is(err, service.ErrArticleForbidden):
		app.Error(c, http.StatusForbidden, errcode.Forbidden, err.Error())
	case errors.Is(err, service.ErrInvalidStatus), errors.Is(err, service.ErrInvalidTags), errors.Is(err, service.ErrInvalidCategory):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, err.Error())
	}
}

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "invalid "+name)
		return 0, false
	}
	return id, true
}

func queryInt(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.Query(key))
	if err != nil {
		return fallback
	}
	return value
}

func queryInt64(c *gin.Context, key string, fallback int64) int64 {
	value, err := strconv.ParseInt(c.Query(key), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
