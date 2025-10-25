# Music Producer Social Network - Backend API

A social media platform for music producers to connect, collaborate, and discover nearby talent through location-based matching.

## 🏗️ Architecture

- **Backend**: Go with Gorilla Mux
- **Database**: PostgreSQL 15 with PostGIS for geospatial queries
- **Cache**: Redis 7 for sessions and caching
- **Storage**: AWS S3 for media files
- **Authentication**: JWT with Redis blacklist
- **Containerization**: Docker & Docker Compose

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- AWS account with S3 bucket (optional for local development)

### Environment Setup

1. Clone the repository and navigate to the backend directory:
```bash
cd backend
```

2. Create a `.env` file with your configuration:
```bash
cp .env.example .env
```

Edit `.env` with your settings:
```env
# Database Configuration
DATABASE_URL=postgres://dev:devpassword@localhost:5432/musicapp?sslmode=disable
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# AWS S3 Configuration (optional for local development)
AWS_ACCESS_KEY_ID=your-aws-access-key-id
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
AWS_REGION=us-east-1
S3_BUCKET_NAME=your-musicapp-media-bucket
S3_CDN_URL=https://your-cdn-domain.com

# Server Configuration
PORT=8080
ENVIRONMENT=development
```

### Running with Docker Compose

1. Start all services:
```bash
docker-compose up -d
```

2. Check service health:
```bash
curl http://localhost:8080/health
```

3. View logs:
```bash
docker-compose logs -f api
```

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start PostgreSQL and Redis:
```bash
docker-compose up -d postgres redis
```

3. Run the API server:
```bash
go run cmd/api/main.go
```

## 📊 Database Schema

The application uses PostgreSQL with PostGIS extension for geospatial queries. Key tables include:

- `users` - User profiles with location data
- `bands` - Band information and location
- `posts` - User and band posts with media support
- `follows` - User following relationships
- `likes` - Post likes
- `reposts` - Post reposts
- `band_members` - Band membership relationships

## 🔐 Authentication

The API uses JWT tokens for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## 📚 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `POST /api/auth/logout` - Logout (blacklist JWT)

### Users
- `GET /api/users/{id}` - Get user profile
- `PUT /api/users/{id}` - Update profile
- `GET /api/users/{id}/posts` - Get user's posts
- `GET /api/users/{id}/followers` - Get followers
- `GET /api/users/{id}/following` - Get following
- `GET /api/users/nearby` - Find nearby users
- `POST /api/users/{id}/profile-picture` - Upload profile picture

### Bands
- `POST /api/bands` - Create band
- `GET /api/bands/{id}` - Get band
- `PUT /api/bands/{id}` - Update band
- `DELETE /api/bands/{id}` - Delete band
- `POST /api/bands/{id}/join` - Join band
- `POST /api/bands/{id}/leave` - Leave band
- `GET /api/bands/{id}/members` - Get band members
- `GET /api/bands/nearby` - Find nearby bands
- `POST /api/bands/{id}/profile-picture` - Upload band profile picture

### Posts
- `POST /api/posts` - Create post
- `GET /api/posts/{id}` - Get post
- `PUT /api/posts/{id}` - Update post
- `DELETE /api/posts/{id}` - Delete post
- `POST /api/posts/{id}/like` - Like post
- `DELETE /api/posts/{id}/like` - Unlike post
- `POST /api/posts/{id}/repost` - Repost
- `POST /api/posts/{id}/media` - Upload media to post

### Social Features
- `POST /api/follow` - Follow user or band
- `DELETE /api/follow` - Unfollow
- `GET /api/feed` - Get personalized feed
- `GET /api/feed/explore` - Get explore feed

## 🗺️ Location-Based Features

The API supports location-based discovery using PostGIS:

### Find Nearby Users
```bash
curl "http://localhost:8080/api/users/nearby?lat=40.7128&lng=-74.0060&radius=50&limit=20"
```

### Find Nearby Bands
```bash
curl "http://localhost:8080/api/bands/nearby?lat=40.7128&lng=-74.0060&radius=50&limit=20"
```

Parameters:
- `lat` - Latitude (required)
- `lng` - Longitude (required)
- `radius` - Search radius in kilometers (default: 50, max: 500)
- `limit` - Maximum results (default: 20, max: 100)

## 📁 Media Upload

The API supports media uploads to AWS S3:

### Supported File Types
- **Images**: JPEG, PNG, GIF (max 5MB)
- **Audio**: MP3, WAV, FLAC (max 50MB)
- **Video**: MP4, MOV, AVI (max 50MB)

### Upload Profile Picture
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -F "profile_picture=@image.jpg" \
  http://localhost:8080/api/users/{id}/profile-picture
```

### Upload Post Media
```bash
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -F "media=@audio.mp3" \
  -F "media=@image.jpg" \
  http://localhost:8080/api/posts/{id}/media
```

## 🔧 Development

### Project Structure
```
backend/
├── cmd/api/           # Application entry point
├── internal/
│   ├── config/        # Configuration management
│   ├── db/           # Database connection
│   ├── cache/        # Redis client
│   ├── middleware/   # HTTP middleware
│   ├── models/       # Data models
│   ├── handlers/     # HTTP handlers
│   └── repository/   # Database queries
├── pkg/utils/        # Shared utilities
├── migrations/       # Database migrations
└── scripts/         # Utility scripts
```

### Running Tests
```bash
go test ./...
```

### Database Migrations
Migrations are automatically applied when starting the PostgreSQL container.

### Health Checks
- `GET /health` - Overall system health
- `GET /ready` - Readiness check

## 🚀 Deployment

### Docker Production Build
```bash
docker build -t musicapp-api .
```

### Environment Variables for Production
- Use strong JWT secrets
- Configure proper CORS origins
- Set up SSL certificates
- Use managed database services
- Configure proper S3 bucket policies

## 📝 API Examples

See [API.md](API.md) for detailed API documentation with example requests and responses.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.
