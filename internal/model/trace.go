package model

type NewTraceData struct {
	TraceId      string
	SessionId    string
	GroupId      string
	ParentId     string
	OwnerId      int
	Name         string
	Status       string
	ErrorMessage string
	StartedAt    int64
	EndedAt      int64
	HasEnded     bool
}

type TraceDTO struct {
	TraceId      string `json:"traceId" validate:"required,uuid"`
	SessionId    string `json:"sessionId" validate:"required,uuid"`
	GroupId      string `json:"groupId" validate:"required,uuid"`
	ParentId     string `json:"parentId" validate:"isdefault|uuid"`
	Name         string `json:"name" validate:"required"`
	Status       string `json:"status" validate:"required,oneof=Ok Error"`
	ErrorMessage string `json:"errorMessage" validate:"required_if=Status Error"`
	StartedAt    int64  `json:"startTime" validate:"required"`
	EndedAt      int64  `json:"endTime" validate:"required"`
	HasEnded     bool   `json:"hasEnded"`
}
