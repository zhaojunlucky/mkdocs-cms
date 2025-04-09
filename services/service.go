package services

import (
	"github.com/zhaojunlucky/mkdocs-cms/core"
)

type Service interface {
	Init(ctx *core.APPContext)
}

type BaseService struct {
	ctx *core.APPContext
}

func (s *BaseService) InitService(name string, ctx *core.APPContext, instance Service) {
	s.ctx = ctx
	s.ctx.RegisterService(name, instance)
}

var service = []Service{
	&SiteService{},
	&UserGitRepoLockService{},
	&AsyncTaskService{},
	&UserGitRepoService{},
	&UserGitRepoCollectionService{},
	&EventService{},
	&UserService{},
}

func InitServices(ctx *core.APPContext) {
	for _, s := range service {
		s.Init(ctx)
	}
}
