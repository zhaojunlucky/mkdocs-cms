package services

import (
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"sync"
)

type UserGitRepoLockService struct {
	BaseService
	repoLocks map[string]*sync.Mutex
	mutex     sync.Mutex
}

func (s *UserGitRepoLockService) Init(ctx *core.APPContext) {
	s.InitService("userGitRepoLockService", ctx, s)
}

func (s *UserGitRepoLockService) Acquire(repoID string) *sync.Mutex {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.repoLocks == nil {
		s.repoLocks = make(map[string]*sync.Mutex)
	}

	if s.repoLocks[repoID] == nil {
		s.repoLocks[repoID] = &sync.Mutex{}
	}
	return s.repoLocks[repoID]
}
