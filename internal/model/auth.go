package model

type NewOwnerData struct {
	Name string
}

type OwnerDTO struct {
	Name string `json:"name" validate:"required"`
}

type NewApiKeyData struct {
	Key     string
	OwnerId int
}

type NewApiKeyDTO struct {
	OwnerId int `param:"ownerId" validate:"required"`
}

type ApiKeyDTO struct {
	Key     string `json:"id" validate:"required"`
	OwnerId int    `json:"ownerId" validate:"required"`
}
