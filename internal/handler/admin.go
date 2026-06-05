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

type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	result, err := h.admin.ListUsers(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 10), c.Query("keyword"))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, result)
}

func (h *AdminHandler) GetDashboard(c *gin.Context) {
	result, err := h.admin.GetDashboard(c.Request.Context())
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, result)
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateUserStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	if err := h.admin.UpdateUserStatus(c.Request.Context(), adminID, userID, req.Status, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	userID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateUserRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	if err := h.admin.UpdateUserRole(c.Request.Context(), adminID, userID, req.Role, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) ListPendingArticles(c *gin.Context) {
	result, err := h.admin.ListPendingArticles(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 10), c.Query("keyword"))
	if err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, result)
}

func (h *AdminHandler) ApproveArticle(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.admin.ApproveArticle(c.Request.Context(), adminID, articleID, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) RejectArticle(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.RejectArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	if err := h.admin.RejectArticle(c.Request.Context(), adminID, articleID, req.Reason, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) ListArticles(c *gin.Context) {
	result, err := h.admin.ListArticles(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 10), service.AdminArticleFilterReq{
		Keyword: c.Query("keyword"), Status: c.Query("status"), UserID: queryInt64(c, "user_id", 0),
		CategoryID: queryInt64(c, "category_id", 0), DateFrom: c.Query("date_from"), DateTo: c.Query("date_to"),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, result)
}

func (h *AdminHandler) DeleteArticle(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.admin.DeleteArticle(c.Request.Context(), adminID, articleID, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) ListComments(c *gin.Context) {
	result, err := h.admin.ListComments(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 10), service.AdminCommentFilterReq{
		Keyword: c.Query("keyword"), UserID: queryInt64(c, "user_id", 0), ArticleID: queryInt64(c, "article_id", 0),
		DateFrom: c.Query("date_from"), DateTo: c.Query("date_to"),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, result)
}

func (h *AdminHandler) DeleteComment(c *gin.Context) {
	adminID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	commentID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.admin.DeleteComment(c.Request.Context(), adminID, commentID, c.ClientIP()); err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, nil)
}

func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	result, err := h.admin.GetAuditLogs(c.Request.Context(), queryInt(c, "page", 1), queryInt(c, "page_size", 10), service.AuditLogFilterReq{
		AdminID: queryInt64(c, "admin_id", 0), Action: c.Query("action"), TargetType: c.Query("target_type"),
		DateFrom: c.Query("date_from"), DateTo: c.Query("date_to"),
	})
	if err != nil {
		writeAdminError(c, err)
		return
	}
	app.Success(c, result)
}

func writeAdminError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidUserStatus), errors.Is(err, service.ErrInvalidUserRole):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrReviewReasonRequired), errors.Is(err, service.ErrInvalidStatus):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrCannotOperateSelf):
		app.Error(c, http.StatusForbidden, errcode.Forbidden, err.Error())
	case errors.Is(err, service.ErrArticleNotFound), errors.Is(err, service.ErrCommentNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
