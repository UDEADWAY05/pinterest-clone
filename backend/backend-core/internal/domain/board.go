package domain

import "time"


type Board struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Title     string    `json:"title"`
	IsPrivate bool      `json:"is_private"`
	CreatedAt time.Time `json:"created_at"`
	IsDefault bool      `json:"is_default"`
}

type BoardPin struct {
	BoardID int64     `json:"board_id"`
	PinID   int64     `json:"pin_id"`
	SavedAt time.Time `json:"saved_at"`
}

type BoardWithRecentPins struct {
	Board
	RecentPins []Pin    `json:"recent_pins"`
}
