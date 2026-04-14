package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type CategoryHandler struct {
	categories *service.CategoryService
}

func NewCategoryHandler(categories *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categories: categories}
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.categories.List(c.Request.Context())
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, categories)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	category, err := h.categories.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryExists) {
			app.Error(c, http.StatusConflict, errcode.InvalidParams, err.Error())
			return
		}
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}

	app.Success(c, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	categoryID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	category, err := h.categories.Update(c.Request.Context(), categoryID, req)
	if err != nil {
		writeCategoryError(c, err)
		return
	}
	app.Success(c, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	categoryID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.categories.Delete(c.Request.Context(), categoryID); err != nil {
		writeCategoryError(c, err)
		return
	}
	app.Success(c, nil)
}

func writeCategoryError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCategoryExists):
		app.Error(c, http.StatusConflict, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrCategoryNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
