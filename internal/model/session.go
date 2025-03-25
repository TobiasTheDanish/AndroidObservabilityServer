package model

type NewSessionDTO struct {
	Id             string
	InstallationId string
	CreatedAt      int64
	Crashed        bool
}
