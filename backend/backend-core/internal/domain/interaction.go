package domain

import "time"

type Interaction struct {
	ID        int64     `json:"id"`
	UserID    *int64    `json:"user_id,omitempty"`
	SessionID string    `json:"session_id"`
	PinID     int64     `json:"pin_id"`
	EventType string    `json:"event_type"`
	TS        time.Time `json:"ts"`
	MetaJSON  string    `json:"meta_json,omitempty"`
}

var AllowedEventTypes = map[string]bool{
	"impression": true, // пин показался в ленте
	"click":      true, // кликнули на пин
	"save":       true, // сохранили в доску
	"unsave":     true, // удалили из доски
	"search":     true, // результат поиска
	"open_pin":   true, // открыли детальный просмотр
	"dwell":      true, // задержались на пине (>3 сек)
}
