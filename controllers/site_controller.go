package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type SiteController struct {
	BaseController
}

func (c *SiteController) Init(ctx *core.APPContext, router *gin.RouterGroup) {
	c.ctx = ctx

	auth := router.Group("/site")
	{
		auth.GET("/version", c.getVersion)
	}
}

func (c *SiteController) getVersion(context *gin.Context) {
	context.JSON(200, gin.H{"version": c.ctx.Version})
}
