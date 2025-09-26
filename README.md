# Dropwise API

Dropwise is a content management and automated delivery system that helps users save, organize, and automatically process their favorite links and content. Think of it as a smart bookmark manager.

## 🚀 Overview

Dropwise allows users to:
- **Save links** with topics, notes, and tags for better organization
- **User management** with secure JWT-based authentication
- **Tag-based organization** for easy content categorization


## 🔧 Technology Stack

- **Backend**: Go 1.24.3 with `net/http` stdlib
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Authentication**: JWT tokens with bcrypt password hashing
- **CORS**: Configured for web frontend integration
- **Migrations**: SQL migrations with goose


...

## 📚 API Endpoints


### Authentication Endpoints

#### Sign Up
```http
POST /api/v1/auth/signup
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2025-06-08T10:00:00Z",
    "updated_at": "2025-06-08T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Sign In
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "created_at": "2025-06-08T10:00:00Z",
    "updated_at": "2025-06-08T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Drops Management Endpoints

> **Note**: All drops endpoints require authentication. Include the JWT token in the Authorization header: `Authorization: Bearer <token>`

#### Create a Drop
```http
POST /api/v1/drops
Authorization: Bearer <token>
Content-Type: application/json

{
  "topic": "Interesting AI Article",
  "url": "https://example.com/ai-article",
  "user_notes": "Great insights on machine learning trends",
  "priority": 5,
  "tags": ["AI", "Machine Learning", "Technology"]
}
```

**Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "topic": "Interesting AI Article",
  "url": "https://example.com/ai-article",
  "user_notes": "Great insights on machine learning trends",
  "added_date": "2025-06-08T10:00:00Z",
  "updated_at": "2025-06-08T10:00:00Z",
  "status": "new",
  "last_sent_date": null,
  "send_count": 0,
  "priority": 5,
  "tags": ["AI", "Machine Learning", "Technology"]
}
```

#### Get All Drops
```http
GET /api/v1/drops
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "topic": "Interesting AI Article",
    "url": "https://example.com/ai-article",
    "user_notes": "Great insights on machine learning trends",
    "added_date": "2025-06-08T10:00:00Z",
    "updated_at": "2025-06-08T10:00:00Z",
    "status": "new",
    "last_sent_date": null,
    "send_count": 0,
    "priority": 5,
    "tags": ["AI", "Machine Learning", "Technology"]
  }
]
```

#### Get Single Drop
```http
GET /api/v1/drops/{id}
Authorization: Bearer <token>
```

#### Update Drop
```http
PUT /api/v1/drops/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "topic": "Updated Topic",
  "user_notes": "Updated notes",
  "priority": 3,
  "status": "sent",
  "tags": ["Updated", "Tags"]
}
```

#### Delete Drop
```http
DELETE /api/v1/drops/{id}
Authorization: Bearer <token>
```

### Tags Endpoints

#### Get All Tags
```http
GET /api/v1/tags
Authorization: Bearer <token>
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "AI"
  },
  {
    "id": 2,
    "name": "Technology"
  }
]
```

### Health Check

#### Server Status
```http
GET /
```

**Response:**
```json
{
  "status": "API is running"
}
```

## 🔐 Authentication

The API uses JWT (JSON Web Tokens) for authentication. After successful login or signup, include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## 📊 Data Models

### Drop
- `id`: Unique identifier (UUID)
- `topic`: Content title/subject
- `url`: Link to the content
- `user_notes`: Personal notes about the content
- `added_date`: When the drop was created
- `updated_at`: Last modification time
- `status`: Processing status (`new`, `sent`, `archived`)
- `last_sent_date`: When it was last processed
- `send_count`: Number of times processed
- `priority`: Processing priority (higher = more important)
- `tags`: Associated tags for organization

### User
- `id`: Unique identifier (UUID)
- `email`: User's email address
- `created_at`: Account creation time
- `updated_at`: Last account update

### Tag
- `id`: Unique identifier
- `name`: Tag name


## 🏷️ Use Cases

**Personal Content Management:**
- Save interesting articles, videos, and resources
- Organize content with tags and notes
- Automatic processing and delivery of saved content
