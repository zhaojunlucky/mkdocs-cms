# MkDocs CMS REST API

A RESTful API for a content management system built with Go and the Gin framework, designed to manage documentation sites.

## Project Structure

```
mkdocs-cms/
├── database/     # Database connection and configuration
├── controllers/  # HTTP request controllers
├── middleware/   # Custom middleware
├── models/       # Data models
├── services/     # Business logic layer
├── main.go       # Application entry point
└── go.mod        # Go module file
```

## Prerequisites

- Go 1.16 or higher
- Git

## Getting Started

1. Clone the repository:
   ```
   git clone https://github.com/zhaojunlucky/mkdocs-cms.git
   cd mkdocs-cms
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Run the application:
   ```
   go run main.go
   ```

4. The API will be available at:
   ```
   http://localhost:8080
   ```

## API Endpoints

### Health Check
- `GET /health` - Check API health status

### Users
- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get a specific user
- `POST /api/v1/users` - Create a new user
- `PUT /api/v1/users/:id` - Update a user
- `DELETE /api/v1/users/:id` - Delete a user

### Posts
- `GET /api/v1/posts` - Get all posts
- `GET /api/v1/posts/:id` - Get a specific post
- `POST /api/v1/posts` - Create a new post
- `PUT /api/v1/posts/:id` - Update a post
- `DELETE /api/v1/posts/:id` - Delete a post

### Git Repositories
- `GET /api/v1/repos` - Get all repositories
- `GET /api/v1/users/:user_id/repos` - Get repositories for a specific user
- `GET /api/v1/repos/:id` - Get a specific repository
- `POST /api/v1/repos` - Create a new repository
- `PUT /api/v1/repos/:id` - Update a repository
- `DELETE /api/v1/repos/:id` - Delete a repository
- `POST /api/v1/repos/:id/sync` - Synchronize a repository with its remote

### Collections
- `GET /api/v1/collections` - Get all collections
- `GET /api/v1/repos/:repo_id/collections` - Get collections for a specific repository
- `GET /api/v1/collections/:id` - Get a specific collection
- `POST /api/v1/collections` - Create a new collection
- `PUT /api/v1/collections/:id` - Update a collection
- `DELETE /api/v1/collections/:id` - Delete a collection
- `GET /api/v1/repos/:repo_id/collections/by-path` - Get a collection by its path

### Events
- `GET /api/v1/events` - Get all events
- `GET /api/v1/events/:id` - Get a specific event
- `GET /api/v1/events/resources/:resource_type` - Get events for a specific resource type
- `POST /api/v1/events` - Create a new event
- `PUT /api/v1/events/:id` - Update an event
- `DELETE /api/v1/events/:id` - Delete an event

### Site Configuration
- `GET /api/v1/site-configs` - Get all site configurations
- `GET /api/v1/repos/:repo_id/site-configs` - Get site configurations for a specific repository
- `GET /api/v1/site-configs/:id` - Get a specific site configuration
- `GET /api/v1/site-configs/by-domain` - Get a site configuration by its domain
- `POST /api/v1/site-configs` - Create a new site configuration
- `PUT /api/v1/site-configs/:id` - Update a site configuration
- `DELETE /api/v1/site-configs/:id` - Delete a site configuration
- `POST /api/v1/site-configs/:id/build` - Update build status of a site
- `POST /api/v1/site-configs/:id/deploy` - Update deployment status of a site

## Request/Response Examples

### Create User

**Request:**
```json
POST /api/v1/users
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepassword",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Response:**
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "created_at": "2025-03-01T10:30:00Z",
  "updated_at": "2025-03-01T10:30:00Z"
}
```

### Create Git Repository

**Request:**
```json
POST /api/v1/repos
{
  "name": "documentation",
  "description": "Project documentation",
  "remote_url": "https://github.com/user/documentation.git",
  "local_path": "/path/to/local/repo",
  "user_id": 1
}
```

**Response:**
```json
{
  "id": 1,
  "name": "documentation",
  "description": "Project documentation",
  "remote_url": "https://github.com/user/documentation.git",
  "local_path": "/path/to/local/repo",
  "user_id": 1,
  "created_at": "2025-03-01T10:40:00Z",
  "updated_at": "2025-03-01T10:40:00Z"
}
```

### Create Site Configuration

**Request:**
```json
POST /api/v1/site-configs
{
  "name": "Project Docs",
  "description": "Project documentation site",
  "repo_working_dir": "/path/to/local/repo/docs",
  "site_domain": "docs.example.com",
  "site_title": "Project Documentation",
  "site_description": "Comprehensive documentation for the project",
  "site_author": "John Doe",
  "theme_name": "material",
  "repo_id": 1
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Project Docs",
  "description": "Project documentation site",
  "repo_working_dir": "/path/to/local/repo/docs",
  "site_domain": "docs.example.com",
  "site_title": "Project Documentation",
  "site_description": "Comprehensive documentation for the project",
  "site_author": "John Doe",
  "site_language": "en",
  "theme_name": "material",
  "repo_id": 1,
  "deployment_method": "github-pages",
  "is_active": true,
  "created_at": "2025-03-01T10:45:00Z",
  "updated_at": "2025-03-01T10:45:00Z"
}
```

## Development

To add new endpoints:
1. Create a new model in the `models` package if needed
2. Implement business logic in the `services` package
3. Create a new controller in the `controllers` package
4. Register the new endpoint in the `setupRoutes` function in `main.go`

## Features

- User management with authentication
- Blog/content post management
- Git repository tracking and synchronization
- Content collection management within repositories
- Comprehensive event logging
- Site configuration management
- Support for multiple deployment methods

## License

This project is licensed under the terms of the license included in the repository.
