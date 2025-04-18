package model

type NewApplicationData struct {
	Name string
}

type ApplicationDTO struct {
	Name   string `json:"name" validate:"required"`
	TeamId int    `param:"teamId" validate:"required"`
}

type NewApiKeyData struct {
	Key   string
	AppId int
}

type NewApiKeyDTO struct {
	AppId int `param:"appId" validate:"required"`
}

type ApiKeyDTO struct {
	Key   string `json:"id" validate:"required"`
	AppId int    `json:"appId" validate:"required"`
}
