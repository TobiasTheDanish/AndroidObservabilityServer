package model

type NewInstallationData struct {
	Id        string
	AppId     int
	Type      string
	Data      map[string]any
	CreatedAt int64
}

type AndroidInstallationDTO struct {
	Id         string `json:"id" validate:"required,uuid"`
	SdkVersion int    `json:"sdkVersion" validate:"required"`
	Model      string `json:"model" validate:"required"`
	Brand      string `json:"brand" validate:"required"`
	CreatedAt  int64  `json:"createdAt" validate:"required"`
}

type InstallationDTO struct {
	Id        string         `json:"id" validate:"required,uuid"`
	Type      string         `param:"type" json:"type" validate:"required"`
	Data      map[string]any `json:"data" validate:"required"`
	CreatedAt int64          `json:"createdAt" validate:"required"`
}

type InstallationEntity struct {
	Id        string
	Type      string
	Data      map[string]any
	CreatedAt int64
	AppId     int
}
