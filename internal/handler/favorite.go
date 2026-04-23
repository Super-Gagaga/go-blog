package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/middleware"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type FavoriteHandler struct {
	favorites *service.FavoriteService
}

func NewFavoriteHandler(favorites *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favorites: favorites}
}

func (h *FavoriteHandler) Toggle(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	articleID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	favorited, err := h.favorites.ToggleFavorite(c.Request.Context(), userID, articleID)
	if err != nil {
		writeInteractionError(c, err)
		return
	}

	app.Success(c, gin.H{"favorited": favorited})
}

func (h *FavoriteHandler) MyFavorites(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		app.Error(c, http.StatusUnauthorized, errcode.Unauthorized, "")
		return
	}

	result, err := h.favorites.GetMyFavorites(
		c.Request.Context(),
		userID,
		queryInt(c, "page", 1),
		queryInt(c, "page_size", 10),
	)
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}

	app.Success(c, result)
}
