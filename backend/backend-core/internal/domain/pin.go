package domain

import "time"

type Pin struct {
	ID          int64   `json:"id" bson:"_id,omitempty"`
	Title       string   `json:"title" bson:"title"`
	Description string   `json:"description" bson:"description"`
	ImageURL    string   `json:"image_url" bson:"image_url"`
	DominantColor string `json:"dominant_color" bson:"dominant_color"`
	Tags        []string `json:"tags" bson:"tags"`
	Width       int      `json:"width" bson:"width"`
	Height      int      `json:"height" bson:"height"`
	AuthorId    int64   `json:"author_id" bson:"author_id"`
	CreatedAt   time.Time    `json:"created_at" bson:"created_at"`
}
