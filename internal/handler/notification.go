package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/middleware"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type NotificationHandler struct {
	notifications *service.NotificationService
}

func NewNotificationHandler(notifications *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifications: notifications}
}

func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	result, err := h.notifications.GetNotifications(c.Request.Context(), userID, queryInt(c, "page", 1), queryInt(c, "page_size", 10))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	unread, err := h.notifications.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		app.Success(c, result)
		return
	}
	app.Success(c, gin.H{"notifications": result, "unread_count": unread})
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}
	var req service.MarkNotificationsReadReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	if err := h.notifications.MarkAsRead(c.Request.Context(), userID, req.IDs); err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, nil)
}
