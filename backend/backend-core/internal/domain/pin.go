package domain

import (
	"time"
)

type Pin struct {
	ID            int64     `json:"id" bson:"_id,omitempty"`
	Title         string    `json:"title" bson:"title"`
	Description   string    `json:"description" bson:"description"`
	ImageURL      string    `json:"image_url" bson:"image_url"`
	DominantColor string    `json:"dominant_color" bson:"dominant_color"`
	Tags          []string  `json:"tags" bson:"tags"`
	Width         int       `json:"width" bson:"width"`
	Height        int       `json:"height" bson:"height"`
	AuthorId      int64     `json:"author_id" bson:"author_id"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	DeletedAt     time.Time `json:"deleted_at" bson:"deleted_at"`
	SavedAt       *time.Time `json:"saved_at,omitempty" bson:"saved_at,omitempty"`
	Author        *UserPreview `json:"author,omitempty" bson:"author,omitempty"`
	SavesCount    int          `json:"saves_count" bson:"saves_count"`
	LikesCount    int          `json:"likes_count" bson:"likes_count"`
	IsSaved       bool         `json:"is_saved" bson:"is_saved"`
}


type UpdatePinParams struct {
	Title       *string
	Description *string
	DominantColor *string
	Tags        *[]string
	ImageURL    *string
	Width       *int
	Height      *int
	AuthorId    *int64
}

type ListPinsResult struct {
	Data       []Pin `json:"data" bson:"data"`
	NextCursor *string `json:"next_cursor" bson:"next_cursor"`
}
