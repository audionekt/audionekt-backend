# Music Producer Social Network - API Documentation

## Base URL
```
http://localhost:8080/api
```

## Authentication

All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Response Format

All API responses follow this format:
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... },
  "error": null
}
```

Error responses:
```json
{
  "success": false,
  "message": null,
  "data": null,
  "error": "Error message"
}
```

---

## Authentication Endpoints

### Register User
**POST** `/auth/register`

Create a new user account.

**Request Body:**
```json
{
  "username": "producer123",
  "email": "producer@example.com",
  "password": "securepassword123",
  "location": {
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "city": "New York",
  "country": "USA"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "producer123",
      "email": "producer@example.com",
      "display_name": null,
      "bio": null,
      "profile_picture_url": null,
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
      },
      "city": "New York",
      "country": "USA",
      "genres": [],
      "skills": [],
      "spotify_url": null,
      "soundcloud_url": null,
      "instagram_handle": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### Login
**POST** `/auth/login`

Authenticate user and get JWT token.

**Request Body:**
```json
{
  "email": "producer@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "username": "producer123",
      "email": "producer@example.com",
      "display_name": "Producer Name",
      "bio": "Music producer from NYC",
      "profile_picture_url": "https://cdn.example.com/profiles/123.jpg",
      "location": {
        "latitude": 40.7128,
        "longitude": -74.0060
      },
      "city": "New York",
      "country": "USA",
      "genres": ["Hip-Hop", "R&B"],
      "skills": ["Mixing", "Mastering"],
      "spotify_url": "https://open.spotify.com/artist/123",
      "soundcloud_url": "https://soundcloud.com/producer123",
      "instagram_handle": "producer123",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

### Logout
**POST** `/auth/logout`

Invalidate JWT token.

**Response:**
```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

---

## User Endpoints

### Get User Profile
**GET** `/users/{id}`

Get user profile by ID.

**Response:**
```json
{
  "success": true,
  "message": "User retrieved successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "username": "producer123",
    "email": "producer@example.com",
    "display_name": "Producer Name",
    "bio": "Music producer from NYC",
    "profile_picture_url": "https://cdn.example.com/profiles/123.jpg",
    "location": {
      "latitude": 40.7128,
      "longitude": -74.0060
    },
    "city": "New York",
    "country": "USA",
    "genres": ["Hip-Hop", "R&B"],
    "skills": ["Mixing", "Mastering"],
    "spotify_url": "https://open.spotify.com/artist/123",
    "soundcloud_url": "https://soundcloud.com/producer123",
    "instagram_handle": "producer123",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Update User Profile
**PUT** `/users/{id}`

Update user profile (requires authentication and ownership).

**Request Body:**
```json
{
  "display_name": "Updated Producer Name",
  "bio": "Updated bio",
  "genres": ["Hip-Hop", "R&B", "Jazz"],
  "skills": ["Mixing", "Mastering", "Beat Making"],
  "spotify_url": "https://open.spotify.com/artist/updated123"
}
```

### Get User Posts
**GET** `/users/{id}/posts?limit=20&offset=0`

Get posts by a specific user.

**Query Parameters:**
- `limit` - Number of posts to return (default: 20, max: 100)
- `offset` - Number of posts to skip (default: 0)

### Get User Followers
**GET** `/users/{id}/followers?limit=20&offset=0`

Get users who follow the specified user.

### Get User Following
**GET** `/users/{id}/following?limit=20&offset=0`

Get users that the specified user is following.

### Find Nearby Users
**GET** `/users/nearby?lat=40.7128&lng=-74.0060&radius=50&limit=20`

Find users within a specified radius.

**Query Parameters:**
- `lat` - Latitude (required)
- `lng` - Longitude (required)
- `radius` - Search radius in kilometers (default: 50, max: 500)
- `limit` - Maximum results (default: 20, max: 100)

### Upload Profile Picture
**POST** `/users/{id}/profile-picture`

Upload a profile picture (requires authentication and ownership).

**Request:** Multipart form data
- `profile_picture` - Image file (JPEG, PNG, GIF, max 5MB)

**Response:**
```json
{
  "success": true,
  "message": "Profile picture uploaded successfully",
  "data": {
    "profile_picture_url": "https://cdn.example.com/profiles/123_profile.jpg"
  }
}
```

---

## Band Endpoints

### Create Band
**POST** `/bands`

Create a new band (requires authentication).

**Request Body:**
```json
{
  "name": "The Producers",
  "bio": "A collective of music producers",
  "location": {
    "latitude": 40.7128,
    "longitude": -74.0060
  },
  "city": "New York",
  "country": "USA",
  "genres": ["Hip-Hop", "R&B"],
  "looking_for": ["Vocalist", "Drummer"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Band created successfully",
  "data": {
    "id": "456e7890-e89b-12d3-a456-426614174000",
    "name": "The Producers",
    "bio": "A collective of music producers",
    "profile_picture_url": null,
    "location": {
      "latitude": 40.7128,
      "longitude": -74.0060
    },
    "city": "New York",
    "country": "USA",
    "genres": ["Hip-Hop", "R&B"],
    "looking_for": ["Vocalist", "Drummer"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Get Band
**GET** `/bands/{id}`

Get band information including members.

### Update Band
**PUT** `/bands/{id}`

Update band details (requires authentication and admin role).

### Delete Band
**DELETE** `/bands/{id}`

Delete a band (requires authentication and admin role).

### Join Band
**POST** `/bands/{id}/join`

Join a band as a member (requires authentication).

### Leave Band
**POST** `/bands/{id}/leave`

Leave a band (requires authentication and membership).

### Get Band Members
**GET** `/bands/{id}/members`

Get all members of a band.

### Find Nearby Bands
**GET** `/bands/nearby?lat=40.7128&lng=-74.0060&radius=50&limit=20`

Find bands within a specified radius.

### Upload Band Profile Picture
**POST** `/bands/{id}/profile-picture`

Upload a band profile picture (requires authentication and admin role).

---

## Post Endpoints

### Create Post
**POST** `/posts`

Create a new post (requires authentication).

**Request Body:**
```json
{
  "content": "Just finished a new beat! Check it out 🎵",
  "media_urls": ["https://cdn.example.com/media/beat1.mp3"],
  "media_types": ["audio"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Post created successfully",
  "data": {
    "id": "789e0123-e89b-12d3-a456-426614174000",
    "author_id": "123e4567-e89b-12d3-a456-426614174000",
    "author_type": "user",
    "band_id": null,
    "user_id": "123e4567-e89b-12d3-a456-426614174000",
    "content": "Just finished a new beat! Check it out 🎵",
    "media_urls": ["https://cdn.example.com/media/beat1.mp3"],
    "media_types": ["audio"],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z",
    "likes_count": 0,
    "reposts_count": 0,
    "is_liked": false,
    "is_reposted": false
  }
}
```

### Get Post
**GET** `/posts/{id}`

Get a specific post with engagement data.

### Update Post
**PUT** `/posts/{id}`

Update a post (requires authentication and ownership).

### Delete Post
**DELETE** `/posts/{id}`

Delete a post (requires authentication and ownership).

### Like Post
**POST** `/posts/{id}/like`

Like a post (requires authentication).

### Unlike Post
**DELETE** `/posts/{id}/like`

Unlike a post (requires authentication).

### Repost
**POST** `/posts/{id}/repost`

Repost a post (requires authentication).

### Upload Media to Post
**POST** `/posts/{id}/media`

Upload media files to a post (requires authentication and ownership).

**Request:** Multipart form data
- `media` - Media files (images, audio, video)

**Response:**
```json
{
  "success": true,
  "message": "Media uploaded successfully",
  "data": {
    "media_urls": [
      "https://cdn.example.com/media/123_audio.mp3",
      "https://cdn.example.com/media/123_image.jpg"
    ],
    "media_types": ["audio", "image"]
  }
}
```

---

## Social Features

### Follow User or Band
**POST** `/follow`

Follow a user or band (requires authentication).

**Request Body:**
```json
{
  "following_type": "user",
  "following_user_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

or

```json
{
  "following_type": "band",
  "following_band_id": "456e7890-e89b-12d3-a456-426614174000"
}
```

### Unfollow
**DELETE** `/follow`

Unfollow a user or band (requires authentication).

**Request Body:** Same as follow request.

### Get Personalized Feed
**GET** `/feed?limit=20&offset=0`

Get posts from followed users and bands (requires authentication).

**Query Parameters:**
- `limit` - Number of posts to return (default: 20, max: 100)
- `offset` - Number of posts to skip (default: 0)

### Get Explore Feed
**GET** `/feed/explore?limit=20&offset=0`

Get trending/explore posts (public endpoint).

---

## Error Codes

- `400` - Bad Request (invalid input)
- `401` - Unauthorized (missing or invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (duplicate resource)
- `500` - Internal Server Error

## Rate Limiting

API endpoints are rate-limited to prevent abuse. Limits are applied per user and per IP address.

## Pagination

Most list endpoints support pagination using `limit` and `offset` parameters:
- `limit` - Maximum number of items to return (default: 20, max: 100)
- `offset` - Number of items to skip (default: 0)

## Media Upload Limits

- **Profile Pictures**: 5MB max (JPEG, PNG, GIF)
- **Post Images**: 10MB max (JPEG, PNG, GIF)
- **Post Audio**: 50MB max (MP3, WAV, FLAC)
- **Post Video**: 50MB max (MP4, MOV, AVI)
