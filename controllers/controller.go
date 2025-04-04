package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type Controller interface {
	Init(ctx *core.APPContext, router *gin.RouterGroup)
}

type BaseController struct {
	ctx *core.APPContext
}

var v1Controllers = []Controller{
	&AsyncTaskController{},
	&UserGitRepoCollectionController{},
	&UserGitRepoController{},
	&GitHubAppController{},
}

var apiControllers = []Controller{
	&GitHubWebhookController{},
	&AuthController{},
}

// InitAPIControllers api
func InitAPIControllers(ctx *core.APPContext, api *gin.RouterGroup) {
	for _, c := range apiControllers {
		c.Init(ctx, api)
	}
}

// InitV1Controllers api/v1
func InitV1Controllers(ctx *core.APPContext, v1 *gin.RouterGroup) {
	for _, c := range v1Controllers {
		c.Init(ctx, v1)
	}
}
