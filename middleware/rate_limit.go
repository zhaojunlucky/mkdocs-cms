package middleware

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zhaojunlucky/mkdocs-cms/core"
	"net/http"
	"regexp"
	"sync"
	"time"
)
import "golang.org/x/time/rate"

type RateLimit struct {
	maxRequests       int
	requestsPerSecond int
	userIPRateLimits  map[string]*rate.Limiter
	mu                sync.Mutex
	ctx               *core.APPContext
}

func (r *RateLimit) CheckRateLimit(c *gin.Context) {
	if r.ShouldSkip(c) {
		c.Next()
		return
	}
	keyVal, exists := c.Get("userId")
	var key string
	if exists {
		key = keyVal.(string)
		log.Info("rate limit userId: ", key)
	}
	if len(key) == 0 {
		key = c.GetHeader("X-Real-IP")
		log.Info("rate limit X-Real-IP: ", key)
	}
	if len(key) == 0 {
		key = c.ClientIP()
		log.Info("rate limit ClientIP: ", key)
	}
	limiter := r.getLimiter(key)
	if !limiter.Allow() {
		core.ResponseErrStr(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
		return
	}
	c.Next()

}

func (r *RateLimit) getLimiter(key string) *rate.Limiter {
	r.mu.Lock()
	defer r.mu.Unlock()
	limiter, ok := r.userIPRateLimits[key]
	if !ok {
		limiter = rate.NewLimiter(rate.Every(time.Second/time.Duration(r.requestsPerSecond)), r.maxRequests)
		r.userIPRateLimits[key] = limiter
	}
	return limiter
}

func (r *RateLimit) ShouldSkip(c *gin.Context) bool {
	skipPaths := []*regexp.Regexp{
		regexp.MustCompile("^/api/auth/logout"),
		regexp.MustCompile("^/api/github/webhook"),
		regexp.MustCompile("^/api/site/version"),
		regexp.MustCompile("^/api/v1/storage/.+"),
	}

	for _, skipPath := range skipPaths {
		if skipPath.MatchString(c.Request.URL.Path) {
			log.Info("rate limit skip path: ", c.Request.URL.Path)
			return true
		}
	}
	return false
}

func NewRateLimit(ctx *core.APPContext) gin.HandlerFunc {
	rateLimit := &RateLimit{
		maxRequests:       ctx.Config.RateLimit.MaxRequests,
		requestsPerSecond: ctx.Config.RateLimit.RequestsPerSecond,
		ctx:               ctx,
		userIPRateLimits:  make(map[string]*rate.Limiter),
	}
	return rateLimit.CheckRateLimit
}
