package models

import (
	"time"
)

// ContentFormat represents the format of content in a collection
type ContentFormat string

const (
	// FormatMarkdown represents Markdown formatted content
	FormatMarkdown ContentFormat = "md"
	// FormatHTML represents HTML formatted content
	FormatHTML ContentFormat = "html"
	// FormatYAML represents YAML formatted content
	FormatYAML ContentFormat = "yaml"
	// FormatJSON represents JSON formatted content
	FormatJSON ContentFormat = "json"
	// FormatText represents plain text content
	FormatText ContentFormat = "txt"
)

type Field struct {
	Type     string `yaml:"type" json:"type"`
	Name     string `yaml:"name" json:"name"`
	Label    string `yaml:"label" json:"label"`
	Required bool   `yaml:"required,omitempty" json:"required"`
	Format   string `yaml:"format,omitempty" json:"format"`
	List     bool   `yaml:"list,omitempty" json:"list"`
}

// UserGitRepoCollection represents a collection of content within a git repository
type UserGitRepoCollection struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	Name        string        `json:"name" gorm:"not null"`
	Label       string        `json:"label" gorm:"not null"`
	Path        string        `json:"path" gorm:"not null"`
	Format      ContentFormat `json:"format" gorm:"type:string;not null;default:'md'"`
	Description string        `json:"description"`
	RepoID      uint          `json:"repo_id" gorm:"not null"`
	Repo        UserGitRepo   `json:"repo" gorm:"foreignKey:RepoID"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Fields      []Field       `json:"fields" gorm:"foreignKey:CollectionID"`
}

// UserGitRepoCollectionResponse is the structure returned to clients
type UserGitRepoCollectionResponse struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Label       string              `json:"label"`
	Path        string              `json:"path"`
	Format      ContentFormat       `json:"format"`
	Description string              `json:"description,omitempty"`
	RepoID      uint                `json:"repo_id"`
	Repo        UserGitRepoResponse `json:"repo,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	Fields      []Field             `json:"fields,omitempty"`
}

// ToResponse converts a UserGitRepoCollection to a UserGitRepoCollectionResponse
func (c *UserGitRepoCollection) ToResponse(includeRepo bool) UserGitRepoCollectionResponse {
	response := UserGitRepoCollectionResponse{
		ID:          c.ID,
		Name:        c.Name,
		Label:       c.Label,
		Path:        c.Path,
		Format:      c.Format,
		Description: c.Description,
		RepoID:      c.RepoID,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Fields:      c.Fields,
	}

	if includeRepo {
		response.Repo = c.Repo.ToResponse(false)
	}

	return response
}

// CreateUserGitRepoCollectionRequest is the structure for collection creation requests
type CreateUserGitRepoCollectionRequest struct {
	Name        string        `json:"name" binding:"required"`
	Label       string        `json:"label" binding:"required"`
	Path        string        `json:"path" binding:"required"`
	Format      ContentFormat `json:"format"`
	Description string        `json:"description"`
	RepoID      uint          `json:"repo_id" binding:"required"`
}

// UpdateUserGitRepoCollectionRequest is the structure for collection update requests
type UpdateUserGitRepoCollectionRequest struct {
	Name        string        `json:"name"`
	Label       string        `json:"label"`
	Path        string        `json:"path"`
	Format      ContentFormat `json:"format"`
	Description string        `json:"description"`
}

// FileUploadRequest represents a request to upload a file
type FileUploadRequest struct {
	Path    string `json:"path" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// FileResponse represents a response with file information
type FileResponse struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"mod_time"`
	Extension string    `json:"extension,omitempty"`
	Content   string    `json:"content,omitempty"`
}
