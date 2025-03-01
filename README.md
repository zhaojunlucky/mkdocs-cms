# MkDocs CMS REST API

A RESTful API for a content management system built with Go and the Gin framework.

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

### Create Post

**Request:**
```json
POST /api/v1/posts
{
  "title": "My First Post",
  "content": "This is the content of my first post.",
  "user_id": 1
}
```

**Response:**
```json
{
  "id": 1,
  "title": "My First Post",
  "content": "This is the content of my first post.",
  "user_id": 1,
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "2025-03-01T10:30:00Z",
    "updated_at": "2025-03-01T10:30:00Z"
  },
  "created_at": "2025-03-01T10:35:00Z",
  "updated_at": "2025-03-01T10:35:00Z"
}
```

## Development

To add new endpoints:
1. Create a new model in the `models` package if needed
2. Implement business logic in the `services` package
3. Create a new controller in the `controllers` package
4. Register the new endpoint in the `setupRoutes` function in `main.go`

## License

This project is licensed under the terms of the license included in the repository.
