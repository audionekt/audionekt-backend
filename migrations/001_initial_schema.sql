-- Enable PostGIS extension
CREATE EXTENSION IF NOT EXISTS postgis;

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100),
    bio TEXT,
    profile_picture_url TEXT,
    location GEOGRAPHY(POINT, 4326), -- PostGIS for lat/long
    city VARCHAR(100),
    country VARCHAR(100),
    genres TEXT[], -- Array of genres (e.g., ['Hip-Hop', 'R&B'])
    skills TEXT[], -- ['Mixing', 'Mastering', 'Beat Making']
    spotify_url TEXT,
    soundcloud_url TEXT,
    instagram_handle VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create spatial index for location queries
CREATE INDEX idx_users_location ON users USING GIST(location);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Bands table
CREATE TABLE bands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    bio TEXT,
    profile_picture_url TEXT,
    location GEOGRAPHY(POINT, 4326),
    city VARCHAR(100),
    country VARCHAR(100),
    genres TEXT[],
    looking_for TEXT[], -- ['Drummer', 'Vocalist', etc.]
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_bands_location ON bands USING GIST(location);
CREATE INDEX idx_bands_name ON bands(name);

-- Band members (many-to-many)
CREATE TABLE band_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    band_id UUID REFERENCES bands(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50), -- 'Producer', 'Vocalist', 'Admin', etc.
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(band_id, user_id)
);

CREATE INDEX idx_band_members_user ON band_members(user_id);
CREATE INDEX idx_band_members_band ON band_members(band_id);

-- Posts table (supports both user and band posts)
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID, -- Can be NULL if band post
    author_type VARCHAR(10) NOT NULL, -- 'user' or 'band'
    band_id UUID REFERENCES bands(id) ON DELETE CASCADE, -- If band post
    user_id UUID REFERENCES users(id) ON DELETE CASCADE, -- If user post
    content TEXT NOT NULL,
    media_urls TEXT[], -- Array of image/audio URLs
    media_types TEXT[], -- ['image', 'audio', 'video']
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT valid_author CHECK (
        (author_type = 'user' AND user_id IS NOT NULL) OR
        (author_type = 'band' AND band_id IS NOT NULL)
    )
);

CREATE INDEX idx_posts_user ON posts(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_posts_band ON posts(band_id) WHERE band_id IS NOT NULL;
CREATE INDEX idx_posts_created ON posts(created_at DESC);
CREATE INDEX idx_posts_author_type ON posts(author_type);

-- Follows table (users follow users and bands)
CREATE TABLE follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_id UUID REFERENCES users(id) ON DELETE CASCADE,
    following_type VARCHAR(10) NOT NULL, -- 'user' or 'band'
    following_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    following_band_id UUID REFERENCES bands(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT valid_following CHECK (
        (following_type = 'user' AND following_user_id IS NOT NULL) OR
        (following_type = 'band' AND following_band_id IS NOT NULL)
    ),
    UNIQUE(follower_id, following_type, following_user_id, following_band_id)
);

CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_user ON follows(following_user_id) WHERE following_user_id IS NOT NULL;
CREATE INDEX idx_follows_band ON follows(following_band_id) WHERE following_band_id IS NOT NULL;

-- Likes table
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, post_id)
);

CREATE INDEX idx_likes_post ON likes(post_id);
CREATE INDEX idx_likes_user ON likes(user_id);

-- Reposts table
CREATE TABLE reposts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, post_id)
);

CREATE INDEX idx_reposts_post ON reposts(post_id);
CREATE INDEX idx_reposts_user ON reposts(user_id);

-- Comments table (optional for Phase 2)
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_user ON comments(user_id);

-- Messages table (for band chat - Phase 2)
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    band_id UUID REFERENCES bands(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_messages_band ON messages(band_id, created_at DESC);
CREATE INDEX idx_messages_user ON messages(user_id);

-- JWT blacklist table for logout functionality
CREATE TABLE jwt_blacklist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jti VARCHAR(255) UNIQUE NOT NULL, -- JWT ID
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_jwt_blacklist_jti ON jwt_blacklist(jti);
CREATE INDEX idx_jwt_blacklist_expires ON jwt_blacklist(expires_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to automatically update updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bands_updated_at BEFORE UPDATE ON bands
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
