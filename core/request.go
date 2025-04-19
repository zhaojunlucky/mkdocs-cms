package core

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strconv"
)

type ParamDef struct {
	name       string
	paramType  string
	empty      bool
	pattern    *regexp.Regexp
	value      any
	statusCode int
	errorMsg   string
}

func (p *ParamDef) Handle(c *gin.Context) error {
	switch p.paramType {
	case "context":
		return p.getContext(c)
	case "query":
		return p.getQuery(c)
	case "url":
		return p.getUrl(c)
	case "cookie":
		return p.getCookie(c)
	default:
		return NewHTTPErrorStr(http.StatusBadRequest, "invalid param type")
	}
}

func (p *ParamDef) String() string {
	return p.value.(string)
}

func (p *ParamDef) Value() any {
	return p.value
}

func (p *ParamDef) UInt64() (uint64, error) {
	repoID, err := strconv.ParseUint(p.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return repoID, nil
}

func (p *ParamDef) Int64() (int64, error) {
	repoID, err := strconv.ParseInt(p.String(), 10, 64)
	if err != nil {
		return 0, err
	}
	return repoID, nil
}

func (p *ParamDef) SetError(code int, errorMsg string) *ParamDef {
	p.statusCode = code
	p.errorMsg = errorMsg
	return p
}

func (p *ParamDef) getContext(c *gin.Context) error {
	errCode := p.statusCode
	if errCode == 0 {
		errCode = http.StatusBadRequest
	}
	errMsg := p.errorMsg
	if errMsg == "" {
		errMsg = fmt.Sprintf("param %s not found or empty", p.name)
	}
	value, exists := c.Get(p.name)
	if !exists {
		return NewHTTPErrorStr(errCode, errMsg)
	}
	if !p.empty && value == "" {
		return NewHTTPErrorStr(errCode, errMsg)
	}

	p.value = value.(string)

	return p.checkPattern()
}

func (p *ParamDef) getQuery(c *gin.Context) error {
	p.value = c.Query(p.name)
	if !p.empty && p.value == "" {
		return NewHTTPErrorStr(http.StatusBadRequest, fmt.Sprintf("param %s is empty", p.name))
	}
	return p.checkPattern()
}

func (p *ParamDef) getUrl(c *gin.Context) error {
	p.value = c.Param(p.name)
	if !p.empty && p.value == "" {
		return NewHTTPErrorStr(http.StatusBadRequest, fmt.Sprintf("param %s is empty", p.name))
	}
	return p.checkPattern()
}

func (p *ParamDef) getCookie(c *gin.Context) error {
	value, err := c.Cookie(p.name)

	if err != nil {
		return NewHTTPErrorStr(http.StatusBadRequest, fmt.Sprintf("param %s not found", p.name))
	}
	p.value = value
	if !p.empty && p.value == "" {
		return NewHTTPErrorStr(http.StatusBadRequest, fmt.Sprintf("param %s is empty", p.name))
	}
	return p.checkPattern()
}

func (p *ParamDef) checkPattern() error {
	if p.pattern != nil {
		// TODO need enhance
		match := p.pattern.MatchString(p.value.(string))
		if !match {
			return NewHTTPErrorStr(http.StatusBadRequest, fmt.Sprintf("param %s does not match pattern %s", p.name, p.pattern.String()))
		}
	}
	return nil
}

type RequestParam struct {
	Params map[string]*ParamDef
}

func NewRequestParam() *RequestParam {
	return &RequestParam{Params: make(map[string]*ParamDef)}
}

func (p *RequestParam) AddContextParam(name string, empty bool, pattern *regexp.Regexp) *ParamDef {
	p.Params[name] = &ParamDef{name: name, empty: empty, pattern: pattern, paramType: "context"}
	return p.Params[name]
}

func (p *RequestParam) AddQueryParam(name string, empty bool, pattern *regexp.Regexp) *ParamDef {
	p.Params[name] = &ParamDef{name: name, empty: empty, pattern: pattern, paramType: "query"}
	return p.Params[name]
}

func (p *RequestParam) AddUrlParam(name string, empty bool, pattern *regexp.Regexp) *ParamDef {
	p.Params[name] = &ParamDef{name: name, empty: empty, pattern: pattern, paramType: "url"}
	return p.Params[name]
}

func (p *RequestParam) AddCookieParam(name string, empty bool, pattern *regexp.Regexp) *ParamDef {
	p.Params[name] = &ParamDef{name: name, empty: empty, pattern: pattern, paramType: "cookie"}
	return p.Params[name]
}

func (p *RequestParam) Handle(c *gin.Context) error {
	for _, param := range p.Params {
		if err := param.Handle(c); err != nil {
			return err
		}
	}
	return nil
}

func (p *RequestParam) HandleWithBody(c *gin.Context, body interface{}) error {
	for _, param := range p.Params {
		if err := param.Handle(c); err != nil {
			return err
		}
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		return NewHTTPErrorStr(http.StatusBadRequest, err.Error())
	}

	return nil
}
