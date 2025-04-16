package model

type NewApplicationData struct {
	Name string
}

type OwnerDTO struct {
	Name string `json:"name" validate:"required"`
}

type NewApiKeyData struct {
	Key   string
	AppId int
}

type NewApiKeyDTO struct {
	AppId int `param:"ownerId" validate:"required"`
}

type ApiKeyDTO struct {
	Key   string `json:"id" validate:"required"`
	AppId int    `json:"ownerId" validate:"required"`
}
