package seo

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type IssueSeverity string

const (
	SeverityInfo    IssueSeverity = "info"
	SeverityWarning IssueSeverity = "warning"
	SeverityError   IssueSeverity = "error"
)

type Issue struct {
	Severity IssueSeverity `json:"severity"`
	Type     string        `json:"type"`
	Message  string        `json:"message"`
	Path     string        `json:"path,omitempty"`
}

type SiteInfo struct {
	SiteName         string `json:"siteName"`
	SiteURL          string `json:"siteUrl"`
	SiteDescription  string `json:"siteDescription"`
	SiteAuthor       string `json:"siteAuthor"`
	UseDirectoryURLs bool   `json:"useDirectoryUrls"`
	DocsDir          string `json:"docsDir"`
	BlogDir          string `json:"blogDir"`
	PostDir          string `json:"postDir"`
	PostURLFormat    string `json:"postUrlFormat"`
}

type Page struct {
	SourcePath   string   `json:"sourcePath"`
	URLPath      string   `json:"urlPath"`
	CanonicalURL string   `json:"canonicalUrl"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	H1           string   `json:"h1"`
	H1Count      int      `json:"h1Count"`
	Tags         []string `json:"tags,omitempty"`
	Draft        bool     `json:"draft,omitempty"`
}

type Report struct {
	Site   SiteInfo `json:"site"`
	Pages  []Page   `json:"pages"`
	Issues []Issue  `json:"issues"`
}

type mkDocsConfig struct {
	SiteName         string        `yaml:"site_name"`
	SiteURL          string        `yaml:"site_url"`
	SiteDescription  string        `yaml:"site_description"`
	SiteAuthor       string        `yaml:"site_author"`
	DocsDir          string        `yaml:"docs_dir"`
	UseDirectoryURLs *bool         `yaml:"use_directory_urls"`
	Plugins          []interface{} `yaml:"plugins"`
}

type blogConfig struct {
	BlogDir       string `yaml:"blog_dir"`
	PostDir       string `yaml:"post_dir"`
	PostURLFormat string `yaml:"post_url_format"`
}

func Scan(root string) (*Report, error) {
	configPath, err := findMkDocsConfig(root)
	if err != nil {
		return nil, err
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config mkDocsConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	site := buildSiteInfo(config)
	report := &Report{Site: site}
	report.Issues = append(report.Issues, siteIssues(root, site)...)

	pages, pageIssues, err := scanPages(root, site)
	if err != nil {
		return nil, err
	}
	report.Pages = pages
	report.Issues = append(report.Issues, pageIssues...)
	report.Issues = append(report.Issues, duplicateIssues(pages)...)

	return report, nil
}

func findMkDocsConfig(root string) (string, error) {
	for _, name := range []string{"mkdocs.yml", "mkdocs.yaml"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", errors.New("mkdocs config not found")
}

func buildSiteInfo(config mkDocsConfig) SiteInfo {
	useDirectoryURLs := true
	if config.UseDirectoryURLs != nil {
		useDirectoryURLs = *config.UseDirectoryURLs
	}

	site := SiteInfo{
		SiteName:         config.SiteName,
		SiteURL:          strings.TrimRight(config.SiteURL, "/"),
		SiteDescription:  config.SiteDescription,
		SiteAuthor:       config.SiteAuthor,
		UseDirectoryURLs: useDirectoryURLs,
		DocsDir:          valueOrDefault(config.DocsDir, "docs"),
		BlogDir:          "blog",
		PostDir:          filepath.ToSlash(filepath.Join("docs", "blog", "Posts")),
		PostURLFormat:    "{date}/{slug}",
	}

	if blog := extractBlogConfig(config.Plugins); blog != nil {
		if blog.BlogDir != "" {
			site.BlogDir = blog.BlogDir
		}
		if blog.PostDir != "" {
			site.PostDir = filepath.ToSlash(blog.PostDir)
		}
		if blog.PostURLFormat != "" {
			site.PostURLFormat = blog.PostURLFormat
		}
	}

	return site
}

func extractBlogConfig(plugins []interface{}) *blogConfig {
	for _, plugin := range plugins {
		pluginMap, ok := plugin.(map[string]interface{})
		if !ok {
			continue
		}
		value, ok := pluginMap["blog"]
		if !ok {
			continue
		}

		data, err := yaml.Marshal(value)
		if err != nil {
			return &blogConfig{}
		}
		var config blogConfig
		_ = yaml.Unmarshal(data, &config)
		return &config
	}
	return nil
}

func siteIssues(root string, site SiteInfo) []Issue {
	var issues []Issue
	if site.SiteName == "" {
		issues = append(issues, Issue{Severity: SeverityWarning, Type: "missing_site_name", Message: "site_name is missing"})
	}
	if site.SiteURL == "" {
		issues = append(issues, Issue{Severity: SeverityError, Type: "missing_site_url", Message: "site_url is missing"})
	}
	if site.SiteDescription == "" {
		issues = append(issues, Issue{Severity: SeverityWarning, Type: "missing_site_description", Message: "site_description is missing"})
	}
	if site.SiteAuthor == "" {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "missing_site_author", Message: "site_author is missing"})
	}

	robotsPath := filepath.Join(root, site.DocsDir, "robots.txt")
	if data, err := os.ReadFile(robotsPath); err == nil && blocksAllRobots(string(data)) {
		issues = append(issues, Issue{Severity: SeverityError, Type: "robots_blocks_all", Path: filepath.ToSlash(filepath.Join(site.DocsDir, "robots.txt")), Message: "robots.txt blocks User-agent: * with Disallow: /"})
	} else if os.IsNotExist(err) {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "missing_robots", Message: "docs/robots.txt is missing"})
	}

	if _, err := os.Stat(filepath.Join(root, "site", "sitemap.xml")); os.IsNotExist(err) {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "sitemap_not_built", Message: "site/sitemap.xml not found; verify mkdocs build generates and publishes sitemap.xml"})
	}

	return issues
}

func scanPages(root string, site SiteInfo) ([]Page, []Issue, error) {
	docsRoot := filepath.Join(root, site.DocsDir)
	var pages []Page
	var issues []Issue

	err := filepath.WalkDir(docsRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(entry.Name()) != ".md" {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sourcePath := filepath.ToSlash(relPath)
		page, pageIssues, err := scanPage(path, sourcePath, site)
		if err != nil {
			return err
		}
		pages = append(pages, page)
		issues = append(issues, pageIssues...)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].SourcePath < pages[j].SourcePath
	})
	return pages, issues, nil
}

func scanPage(path string, sourcePath string, site SiteInfo) (Page, []Issue, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Page{}, nil, err
	}

	frontMatter, body := splitFrontMatter(data)
	meta := map[string]interface{}{}
	if len(frontMatter) > 0 {
		_ = yaml.Unmarshal(frontMatter, &meta)
	}

	title := stringValue(meta["title"])
	description := stringValue(meta["description"])
	h1 := firstH1(body)
	h1Count := countH1(body)
	page := Page{
		SourcePath:  sourcePath,
		URLPath:     BuildURLPath(sourcePath, meta, site),
		Title:       title,
		Description: description,
		H1:          h1,
		H1Count:     h1Count,
		Tags:        stringSlice(meta["tags"]),
		Draft:       boolValue(meta["draft"]),
	}
	page.CanonicalURL = joinURL(site.SiteURL, page.URLPath)

	return page, pageIssues(page), nil
}

func pageIssues(page Page) []Issue {
	var issues []Issue
	if page.Draft {
		return issues
	}
	if page.Title == "" {
		issues = append(issues, Issue{Severity: SeverityWarning, Type: "missing_title", Path: page.SourcePath, Message: "page title is missing"})
	}
	if page.Description == "" {
		issues = append(issues, Issue{Severity: SeverityWarning, Type: "missing_description", Path: page.SourcePath, Message: "page description is missing"})
	} else if len([]rune(page.Description)) < 80 {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "short_description", Path: page.SourcePath, Message: "page description is shorter than 80 characters"})
	} else if len([]rune(page.Description)) > 160 {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "long_description", Path: page.SourcePath, Message: "page description is longer than 160 characters"})
	}
	if page.H1 == "" {
		issues = append(issues, Issue{Severity: SeverityWarning, Type: "missing_h1", Path: page.SourcePath, Message: "page H1 is missing"})
	}
	if page.H1Count > 1 {
		issues = append(issues, Issue{Severity: SeverityInfo, Type: "multiple_h1", Path: page.SourcePath, Message: "page has more than one H1"})
	}
	return issues
}

func duplicateIssues(pages []Page) []Issue {
	var issues []Issue
	issues = append(issues, duplicateValueIssues(pages, "title")...)
	issues = append(issues, duplicateValueIssues(pages, "description")...)
	return issues
}

func duplicateValueIssues(pages []Page, field string) []Issue {
	seen := map[string][]string{}
	for _, page := range pages {
		var value string
		if field == "title" {
			value = page.Title
		} else {
			value = page.Description
		}
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" || page.Draft {
			continue
		}
		seen[value] = append(seen[value], page.SourcePath)
	}

	var issues []Issue
	for _, paths := range seen {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			issues = append(issues, Issue{Severity: SeverityInfo, Type: "duplicate_" + field, Path: path, Message: fmt.Sprintf("duplicate page %s", field)})
		}
	}
	return issues
}

func BuildURLPath(sourcePath string, meta map[string]interface{}, site SiteInfo) string {
	sourcePath = filepath.ToSlash(sourcePath)
	postDir := filepath.ToSlash(site.PostDir)
	if postDir != "" && strings.HasPrefix(sourcePath, strings.TrimRight(postDir, "/")+"/") {
		return buildBlogURLPath(sourcePath, meta, site)
	}

	docsPrefix := strings.TrimRight(filepath.ToSlash(site.DocsDir), "/") + "/"
	path := strings.TrimPrefix(sourcePath, docsPrefix)
	path = strings.TrimSuffix(path, filepath.Ext(path))
	if strings.HasSuffix(path, "/index") {
		path = strings.TrimSuffix(path, "/index")
	}
	if site.UseDirectoryURLs {
		return "/" + strings.Trim(path, "/") + "/"
	}
	return "/" + strings.Trim(path, "/") + ".html"
}

func buildBlogURLPath(sourcePath string, meta map[string]interface{}, site SiteInfo) string {
	name := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	date, slug := splitDatedSlug(name)
	if metaDate := dateValueFromMeta(meta["date"]); metaDate != "" {
		date = metaDate
	}
	if slug == "" {
		slug = slugify(name)
	}

	path := site.PostURLFormat
	path = strings.ReplaceAll(path, "{date}", date)
	path = strings.ReplaceAll(path, "{slug}", slug)
	return "/" + strings.Trim(site.BlogDir, "/") + "/" + strings.Trim(path, "/") + "/"
}

func splitDatedSlug(name string) (string, string) {
	re := regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(.+)$`)
	matches := re.FindStringSubmatch(name)
	if len(matches) == 3 {
		return matches[1], slugify(matches[2])
	}
	return "", slugify(name)
}

func splitFrontMatter(data []byte) ([]byte, []byte) {
	if !bytes.HasPrefix(data, []byte("---")) {
		return nil, data
	}
	closeIdx := bytes.Index(data[3:], []byte("\n---"))
	if closeIdx < 0 {
		return nil, data
	}
	closeIdx += 3
	end := bytes.IndexByte(data[closeIdx+1:], '\n')
	if end < 0 {
		return data, nil
	}
	end += closeIdx + 2
	return data[4 : closeIdx+1], data[end:]
}

func firstH1(body []byte) string {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func countH1(body []byte) int {
	count := 0
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			count++
		}
	}
	return count
}

func blocksAllRobots(content string) bool {
	currentApplies := false
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(strings.Split(line, "#")[0])
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		switch key {
		case "user-agent":
			currentApplies = value == "*"
		case "disallow":
			if currentApplies && value == "/" {
				return true
			}
		}
	}
	return false
}

func stringValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func stringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	case []string:
		return v
	default:
		return nil
	}
}

func boolValue(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	default:
		return false
	}
}

func dateValueFromMeta(value interface{}) string {
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		if len(v) >= 10 {
			return v[:10]
		}
	}
	return ""
}

func slugify(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func joinURL(siteURL string, path string) string {
	if siteURL == "" {
		return path
	}
	parsed, err := url.Parse(siteURL)
	if err != nil {
		return strings.TrimRight(siteURL, "/") + path
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed.String()
}

func valueOrDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
