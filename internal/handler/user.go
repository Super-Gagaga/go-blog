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

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	user, err := h.users.Register(c.Request.Context(), req)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	tokens, err := h.users.Login(c.Request.Context(), req)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, tokens)
}

func (h *UserHandler) Refresh(c *gin.Context) {
	var req service.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	tokens, err := h.users.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, tokens)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	profile, err := h.users.GetProfile(c.Request.Context(), userID)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	var req service.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	profile, err := h.users.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, profile)
}

func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "avatar is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "open avatar failed")
		return
	}
	defer file.Close()

	url, err := h.users.UploadAvatar(c.Request.Context(), userID, file, fileHeader.Filename, fileHeader.Size)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, gin.H{"avatar": url})
}

func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || userID <= 0 {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, "invalid user id")
		return
	}

	profile, err := h.users.GetPublicProfile(c.Request.Context(), userID)
	if err != nil {
		writeUserError(c, err)
		return
	}

	app.Success(c, profile)
}

func writeUserError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserExists):
		app.Error(c, http.StatusConflict, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrInvalidCredential), errors.Is(err, service.ErrInvalidRefresh):
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, err.Error())
	case errors.Is(err, service.ErrUserBanned):
		app.Error(c, http.StatusForbidden, errcode.Forbidden, err.Error())
	case errors.Is(err, service.ErrUserNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	case errors.Is(err, service.ErrUnsupportedAvatar), errors.Is(err, service.ErrAvatarTooLarge):
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
