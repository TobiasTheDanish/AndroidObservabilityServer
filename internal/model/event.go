package model

type NewEventData struct {
	Id             string
	SessionId      string
	AppId          int
	Type           string
	SerializedData string
	CreatedAt      int64
}

type EventEntity struct {
	Id             string
	SessionId      string
	AppId          int
	Type           string
	SerializedData string
	CreatedAt      int64
}

type EventDTO struct {
	Id             string `json:"id" validate:"required,uuid"`
	SessionId      string `json:"sessionId" validate:"required,uuid"`
	Type           string `json:"type" validate:"required"`
	SerializedData string `json:"serializedData"`
	CreatedAt      int64  `json:"createdAt" validate:"required"`
}
