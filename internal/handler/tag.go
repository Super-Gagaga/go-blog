package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/pkg/app"
	"github.com/yourname/go-blog/internal/pkg/errcode"
	"github.com/yourname/go-blog/internal/service"
)

type TagHandler struct {
	tags *service.TagService
}

func NewTagHandler(tags *service.TagService) *TagHandler {
	return &TagHandler{tags: tags}
}

func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.tags.List(c.Request.Context(), c.Query("keyword"))
	if err != nil {
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}
	app.Success(c, tags)
}

func (h *TagHandler) Create(c *gin.Context) {
	var req service.CreateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}

	tag, err := h.tags.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrTagExists) {
			app.Error(c, http.StatusConflict, errcode.InvalidParams, err.Error())
			return
		}
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
		return
	}

	app.Success(c, tag)
}

func (h *TagHandler) Update(c *gin.Context) {
	tagID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req service.UpdateTagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Error(c, http.StatusBadRequest, errcode.InvalidParams, err.Error())
		return
	}
	tag, err := h.tags.Update(c.Request.Context(), tagID, req)
	if err != nil {
		writeTagError(c, err)
		return
	}
	app.Success(c, tag)
}

func (h *TagHandler) Delete(c *gin.Context) {
	tagID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.tags.Delete(c.Request.Context(), tagID); err != nil {
		writeTagError(c, err)
		return
	}
	app.Success(c, nil)
}

func writeTagError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTagExists):
		app.Error(c, http.StatusConflict, errcode.InvalidParams, err.Error())
	case errors.Is(err, service.ErrTagNotFound):
		app.Error(c, http.StatusNotFound, errcode.NotFound, err.Error())
	default:
		app.Error(c, http.StatusInternalServerError, errcode.InternalServer, "")
	}
}
