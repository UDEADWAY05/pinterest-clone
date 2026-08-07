package domain

import "time"

type Comment struct {
	ID        int64     `json:"id" bson:"id"`
	PinID     int64     `json:"pin_id" bson:"pin_id"`
	AuthorID  int64     `json:"author_id" bson:"author_id"`
	Body      string    `json:"body" bson:"body"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type ListCommentsResult struct {
	Data       []Comment `json:"data" bson:"data"`
	NextCursor *string   `json:"next_cursor" bson:"next_cursor"`
}
