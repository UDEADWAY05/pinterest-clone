package domain

import "time"

type User struct {
	ID           int64     `json:"id" bson:"_id,omitempty"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name" bson:"display_name"`
	AvatarURL    string    `json:"avatar_url" bson:"avatar_url"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	Favorites    []int64   `json:"favorites" bson:"favorites"`
}

type UserProfile struct {
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	DisplayName    string    `json:"display_name"`
	AvatarURL      string    `json:"avatar_url"`
	CreatedAt      time.Time `json:"created_at"`
	PinsCount      int       `json:"pins_count"`
	FollowersCount int       `json:"followers_count"`
	FollowingCount int       `json:"following_count"`
	IsFollowing    bool      `json:"is_following"`
}

// id BIGSERIAL PRIMARY KEY,
// email TEXT NOT NULL UNIQUE,
// password_hash TEXT NOT NULL,
// display_name TEXT NOT NULL,
// avatar_url TEXT,
// created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
