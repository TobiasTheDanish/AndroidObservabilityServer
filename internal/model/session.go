package model

type NewSessionData struct {
	Id             string
	InstallationId string
	AppId          int
	CreatedAt      int64
	Crashed        bool
}

type SessionDTO struct {
	Id             string `json:"id" validate:"required,uuid"`
	InstallationId string `json:"installationId" validate:"required"`
	CreatedAt      int64  `json:"createdAt" validate:"required"`
	Crashed        bool   `json:"crashed"`
}

type SessionEntity struct {
	Id             string
	InstallationId string
	CreatedAt      int64
	Crashed        bool
	AppId          int
}
