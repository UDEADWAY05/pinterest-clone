package domain

import "time"

type Follow struct {
	FollowerID  int64     `json:"follower_id" bson:"follower_id"`
	FollowingID int64     `json:"following_id" bson:"following_id"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

type ListFollowsResult struct {
	Data       []Follow `json:"data" bson:"data"`
	NextCursor *string  `json:"next_cursor" bson:"next_cursor"`
}
