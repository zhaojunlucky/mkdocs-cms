package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zhaojunlucky/mkdocs-cms/models"
)

func TestResolveConfiguredEditPathChoosesLongestCollectionRoot(t *testing.T) {
	repo := newTestRepo(t)
	service := &UserGitRepoCollectionService{}

	resolved, err := service.ResolveConfiguredEditPath(repo, "blog/Posts/2026/2026-08-17-kubectl-context.md")
	if err != nil {
		t.Fatalf("ResolveConfiguredEditPath returned error: %v", err)
	}
	if resolved.CollectionName != "post" {
		t.Fatalf("collection = %q, want %q", resolved.CollectionName, "post")
	}
	if resolved.Path != "2026/2026-08-17-kubectl-context.md" {
		t.Fatalf("path = %q, want %q", resolved.Path, "2026/2026-08-17-kubectl-context.md")
	}
}

func TestResolveConfiguredEditPathRejectsUnconfiguredPath(t *testing.T) {
	repo := newTestRepo(t)
	service := &UserGitRepoCollectionService{}

	if _, err := service.ResolveConfiguredEditPath(repo, "2026/2026-08-17-kubectl-context.md"); err == nil {
		t.Fatal("ResolveConfiguredEditPath returned nil error for path outside configured collection roots")
	}
}

func TestGetFileContentKeepsCollectionRelativePath(t *testing.T) {
	repo := newTestRepo(t)
	service := &UserGitRepoCollectionService{}

	content, contentType, err := service.GetFileContent(repo, "post", "2026/2026-08-17-kubectl-context.md")
	if err != nil {
		t.Fatalf("GetFileContent returned error: %v", err)
	}
	if string(content) != "# Kubectl Context\n" {
		t.Fatalf("content = %q, want %q", string(content), "# Kubectl Context\n")
	}
	if contentType != "text/markdown" {
		t.Fatalf("contentType = %q, want %q", contentType, "text/markdown")
	}
}

func TestGetFileContentDoesNotResolveMkDocsEditPath(t *testing.T) {
	repo := newTestRepo(t)
	service := &UserGitRepoCollectionService{}

	_, _, err := service.GetFileContent(repo, "post", "blog/Posts/2026/2026-08-17-kubectl-context.md")
	if err == nil {
		t.Fatal("GetFileContent returned nil error for MkDocs edit path")
	}
}

func TestGetFileContentRejectsTraversal(t *testing.T) {
	repo := newTestRepo(t)
	service := &UserGitRepoCollectionService{}

	if _, _, err := service.GetFileContent(repo, "post", "blog/Posts/../secret.md"); err == nil {
		t.Fatal("GetFileContent returned nil error for traversal path")
	}
}

func newTestRepo(t *testing.T) *models.UserGitRepo {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "veda"), 0755); err != nil {
		t.Fatalf("mkdir veda: %v", err)
	}
	config := `collections:
  - name: exia
    label: Exia
    path: docs
    format: md
  - name: post
    label: Posts
    path: docs/blog/Posts
    format: md
`
	if err := os.WriteFile(filepath.Join(root, "veda", "config.yml"), []byte(config), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	writeTestFile(t, root, "docs/blog/Posts/2026/2026-08-17-kubectl-context.md")

	return &models.UserGitRepo{
		ID:        1,
		Name:      "repo",
		LocalPath: root,
	}
}

func writeTestFile(t *testing.T, root string, path string) {
	t.Helper()

	fullPath := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir file parent: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("# Kubectl Context\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
