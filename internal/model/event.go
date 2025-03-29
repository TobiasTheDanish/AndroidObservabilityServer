package model

type NewEventData struct {
	ID             string
	SessionID      string
	Type           string
	SerializedData string
	CreatedAt      int64
}

type EventDTO struct {
	ID             string `json:"id" validate:"required,uuid"`
	SessionID      string `json:"sessionId" validate:"required,uuid"`
	Type           string `json:"type" validate:"required"`
	SerializedData string `json:"serializedData"`
	CreatedAt      int64  `json:"createdAt" validate:"required"`
}
