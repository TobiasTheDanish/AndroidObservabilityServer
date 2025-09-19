package model

type CreateLogDTO struct {
	SessionId string         `json:"sessionId" validate:"required,uuid"`
	Message   string         `json:"message" validate:"required"`
	Data      map[string]any `json:"data"`
	CreatedAt int64          `json:"createdAt" validate:"required"`
}

type GetLogDTO struct {
	Id        int            `json:"id" validate:"required"`
	SessionId string         `json:"sessionId" validate:"required,uuid"`
	Message   string         `json:"message" validate:"required"`
	Data      map[string]any `json:"data"`
	CreatedAt int64          `json:"createdAt" validate:"required"`
}

type NewLogData struct {
	AppId     int
	SessionId string
	Message   string
	Data      map[string]any
	CreatedAt int64
}

type LogEntity struct {
	Id        int
	AppId     int
	SessionId string
	Message   string
	Data      map[string]any
	CreatedAt int64
}
