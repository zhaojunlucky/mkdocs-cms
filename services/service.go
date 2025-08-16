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
	&MetricsService{},
	&MinIOService{},
	&SiteService{},
	&UserFileDraftStatusService{},
	&UserService{},
	&StorageService{},
	&UserGitRepoLockService{},
	&AsyncTaskService{},
	&UserGitRepoService{},
	&UserGitRepoCollectionService{},
	&EventService{},
}

func InitServices(ctx *core.APPContext) {
	for _, s := range service {
		s.Init(ctx)
	}
}
