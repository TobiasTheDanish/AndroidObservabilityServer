package model

type NewEventDTO struct {
	ID             string
	SessionID      string
	Type           string
	SerializedData string
	CreatedAt      int64
}
