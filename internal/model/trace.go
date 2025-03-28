package model

type NewTraceDTO struct {
	TraceId      string
	SessionId    string
	GroupId      string
	ParentId     string
	Name         string
	Status       string
	ErrorMessage string
	StartedAt    int64
	EndedAt      int64
	HasEnded     bool
}
