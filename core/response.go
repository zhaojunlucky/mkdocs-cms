package core

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func HandleError(c *gin.Context, err error) {
	if err == nil {
		panic("unreachable")
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func EnsureNonNilArr[T any](arr []T) []T {
	if arr == nil {
		return make([]T, 0)
	}
	return arr
}
