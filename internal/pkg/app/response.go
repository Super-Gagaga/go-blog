package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yourname/go-blog/internal/model"
	"github.com/yourname/go-blog/internal/pkg/errcode"
)

func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, errcode.Success, errcode.Message(errcode.Success), data)
}

func SuccessMessage(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, errcode.Success, message, data)
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
	if message == "" {
		message = errcode.Message(code)
	}
	JSON(c, httpStatus, code, message, nil)
}

func JSON(c *gin.Context, httpStatus int, code int, message string, data interface{}) {
	c.JSON(httpStatus, model.Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
