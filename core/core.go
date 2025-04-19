package core

import (
	"errors"
	"fmt"
	"github.com/google/go-github/v45/github"
	"github.com/zhaojunlucky/mkdocs-cms/config"
	"github.com/zhaojunlucky/mkdocs-cms/models"
	"gorm.io/gorm"
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTPError: StatusCode=%d, Message=%s", e.StatusCode, e.Message)
}

func NewHTTPErrorStr(code int, message string) error {
	return &HTTPError{
		StatusCode: code,
		Message:    message,
	}
}

func NewHTTPError(code int, err error) error {
	return &HTTPError{
		StatusCode: code,
		Message:    err.Error(),
	}
}

func NewGormHTTPError(err error) *HTTPError {

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &HTTPError{
			StatusCode: 404,
			Message:    err.Error(),
		}
	}

	return &HTTPError{
		StatusCode: 500,
		Message:    err.Error(),
	}
}

type APPContext struct {
	GithubAppSettings *models.GitHubAppSettings
	ServiceMap        map[string]interface{}
	Config            *config.Config
	GithubAppClient   *github.Client
	RepoBasePath      string
	LogDirPath        string
	Version           string
	CookieDomain      string
}

func (c *APPContext) RegisterService(name string, service interface{}) {
	if c.ServiceMap == nil {
		c.ServiceMap = make(map[string]interface{})
	}
	c.ServiceMap[name] = service
}

func (c *APPContext) MustGetService(name string) interface{} {
	if c.ServiceMap[name] == nil {
		panic(fmt.Sprintf("service %s not found", name))
	}
	return c.ServiceMap[name]
}
