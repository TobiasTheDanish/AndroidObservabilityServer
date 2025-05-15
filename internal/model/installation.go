package model

type NewInstallationData struct {
	Id         string
	AppId      int
	SdkVersion int
	Model      string
	Brand      string
	CreatedAt  int64
}

type InstallationDTO struct {
	Id         string `json:"id" validate:"required,uuid"`
	SdkVersion int    `json:"sdkVersion" validate:"required"`
	Model      string `json:"model" validate:"required"`
	Brand      string `json:"brand" validate:"required"`
	CreatedAt  int64  `json:"createdAt" validate:"required"`
}

type InstallationEntity struct {
	Id         string
	SDKVersion int
	Model      string
	Brand      string
	CreatedAt  int64
	AppId      int
}
