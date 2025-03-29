package core

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ArrayResponse[T any] struct {
	Data       []T `json:"entries"`
	EntryCount int `json:"entryCount"`
}

func NewArrayResponse[T any](data []T) ArrayResponse[T] {
	return ArrayResponse[T]{
		Data:       EnsureNonNilArr(data),
		EntryCount: len(data),
	}
}

func HandleError(c *gin.Context, err error) {
	if err == nil {
		panic("unreachable")
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.StatusCode, NewErrorMessageDTO(httpErr.StatusCode, err))
	} else {
		c.JSON(http.StatusInternalServerError, NewErrorMessageDTO(http.StatusInternalServerError, err))
	}
}

func EnsureNonNilArr[T any](arr []T) []T {
	if arr == nil {
		return make([]T, 0)
	}
	return arr
}

func ResponseErr(c *gin.Context, httpStatusCode int, err ...error) {
	c.JSON(httpStatusCode, NewErrorMessageDTO(httpStatusCode, err...))
}

func ResponseErrStr(c *gin.Context, httpStatusCode int, message ...string) {
	c.JSON(httpStatusCode, NewErrorMessageDTOStr(httpStatusCode, message...))
}

func ResponseOKArr[T any](c *gin.Context, data []T) {
	ResponseArr(c, http.StatusOK, data)
}

func ResponseCreatedArr[T any](c *gin.Context, data []T) {
	ResponseArr(c, http.StatusCreated, data)
}

func ResponseArr[T any](c *gin.Context, statusCode int, data []T) {
	c.JSON(statusCode, NewArrayResponse(data))
}
