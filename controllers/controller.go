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

var rootControllers = []Controller{
	&MetricsController{},
}

var v1Controllers = []Controller{
	&AsyncTaskController{},
	&UserGitRepoCollectionController{},
	&UserGitRepoController{},
	&GitHubAppController{},
	&StorageController{},
}

var apiControllers = []Controller{
	&GitHubWebhookController{},
	&AuthController{},
	&SiteController{},
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

func InitRootController(ctx *core.APPContext, root *gin.RouterGroup) {
	for _, c := range rootControllers {
		c.Init(ctx, root)
	}
}
