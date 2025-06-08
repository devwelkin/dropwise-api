# Dropwise API

Dropwise is a content management and automated delivery system that helps users save, organize, and automatically process their favorite links and content. Think of it as a smart bookmark manager.

## 🚀 Overview

Dropwise allows users to:
- **Save links** with topics, notes, and tags for better organization
- **Automatic processing** of saved content through a background worker system
- **User management** with secure JWT-based authentication
- **Tag-based organization** for easy content categorization
- **Priority-based content delivery** system

## 🏗️ Architecture

The application consists of two main components:
- **API Server** (`cmd/api`): REST API for user interactions
- **Worker Process** (`cmd/worker`): Background service for notification

## 🔧 Technology Stack

- **Backend**: Go 1.24.3 with `net/http` stdlib
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Authentication**: JWT tokens with bcrypt password hashing
- **CORS**: Configured for web frontend integration
- **Migrations**: SQL migrations with goose

## ✅ Current Implementation Status

### ✅ Completed Features

#### **Authentication System**
- ✅ User registration with email/password
- ✅ Secure login with JWT tokens
- ✅ Password hashing with bcrypt
- ✅ Protected route middleware

#### **Drops CRUD Operations**
- ✅ **CREATE**: Add new drops with topic, URL, notes, priority, and tags
- ✅ **READ**: List all user drops & fetch individual drops by ID
- ✅ **UPDATE**: Modify drop properties including topic, URL, notes, priority, status, and tags
- ✅ **DELETE**: Remove drops with proper user authorization

#### **Tags Management**
- ✅ Create tags automatically when adding drops
- ✅ Associate multiple tags with drops
- ✅ List all available tags
- ✅ Update tags when modifying drops

#### **Background Worker System** 🚧
- ✅ Basic worker infrastructure setup
- ✅ Drop processing logic framework
- ✅ Priority-based processing (higher priority first)
- ✅ One drop per user per cycle
- ✅ Status tracking (`new` → `sent`)
- ✅ Send count and last sent date tracking
- ✅ Cloud Function deployment ready
- ⏳ **In Progress**: Actual notification delivery (email, push, etc.)
- ⏳ **In Progress**: Reminder scheduling system
- ⏳ **In Progress**: Smart timing algorithms

#### **Database Design**
- ✅ PostgreSQL with UUID primary keys
- ✅ Type-safe queries with SQLC
- ✅ Database migrations with Goose
- ✅ Proper indexing for performance
- ✅ Referential integrity with foreign keys

## 🔮 Roadmap: Smart Reminder System

### 🎯 Phase 1: Advanced Scheduling (Q1 2025)
- [ ] **Custom Reminder Intervals**: Allow users to set custom intervals (daily, weekly, monthly, custom)
- [ ] **Scheduled Reminders**: Time-based reminders (morning, afternoon, evening)
- [ ] **Smart Snoozing**: Intelligent snooze functionality with multiple snooze options
- [ ] **Reminder Templates**: Pre-defined reminder patterns for different content types

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

## 🤖 Worker System

The application includes a background worker system foundation that:
- ✅ **Infrastructure**: Basic worker setup with database integration
- ✅ **Processing Logic**: Fetches drops with `new` status for processing
- ✅ **Priority Handling**: Processes drops by priority level and creation date
- ✅ **User Management**: Handles one drop per user per cycle
- ✅ **Status Tracking**: Updates drop status and tracking information
- ✅ **Deployment Ready**: Can be run as standalone process or Cloud Function
- 🚧 **In Development**: Actual notification delivery mechanisms
- 🚧 **Planned**: Email, push notifications, and smart scheduling

**Current State**: The worker simulates sending notifications but doesn't yet deliver actual reminders. The infrastructure is ready for implementing various notification channels.

## 🚦 Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or invalid
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## 🔧 Configuration

The application uses environment variables for configuration:
- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: Secret key for JWT token signing


## 🌐 CORS Configuration

The API is configured to accept requests from:
- `https://dropwise.vercel.app` (Production frontend)
- `http://localhost:5173` (Development frontend)

## 🏷️ Use Cases

**Personal Content Management:**
- Save interesting articles, videos, and resources
- Organize content with tags and notes
- Automatic processing and delivery of saved content
