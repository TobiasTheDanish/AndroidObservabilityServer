package model

type NewApplicationData struct {
	Name   string
	TeamId int
}

type ApplicationEntity struct {
	Id     int
	Name   string
	TeamId int
}

type CreateApplicationDTO struct {
	Name   string `json:"name" validate:"required"`
	TeamId int    `json:"teamId" validate:"required"`
}

type GetApplicationDTO struct {
	Id     int    `json:"id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	TeamId int    `json:"teamId" validate:"required"`
}

type ApplicationDataEntity struct {
	Installations []InstallationEntity
	Sessions      []SessionEntity
}

type ApplicationDataDTO struct {
	Installations []InstallationDTO `json:"installations"`
	Sessions      []SessionDTO      `json:"sessions"`
}
