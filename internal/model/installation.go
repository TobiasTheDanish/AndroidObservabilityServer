package model

type NewInstallationData struct {
	Id         string
	OwnerId    int
	SdkVersion int
	Model      string
	Brand      string
}

type InstallationDTO struct {
	Id         string `json:"id" validate:"required,uuid"`
	SdkVersion int    `json:"sdkVersion" validate:"required"`
	Model      string `json:"model" validate:"required"`
	Brand      string `json:"brand" validate:"required"`
}
