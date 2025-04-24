package model

type SignInDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type NewAuthSessionData struct {
	Id     string
	UserId int
	Expiry int64
}

type AuthSessionEntity struct {
	Id     string
	UserId int
	Expiry int64
}

type NewTeamData struct {
	Name string
}

type TeamEntity struct {
	Id   int
	Name string
}

type GetTeamDTO struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type CreateTeamDTO struct {
	Name string `json:"name" validate:"required"`
}

type NewUserData struct {
	Name         string
	PasswordHash string
}

type UserEntity struct {
	Id           int
	Name         string
	PasswordHash string
}

type UserDTO struct {
	Name     string `json:"name" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type NewTeamUserLinkData struct {
	TeamId int
	UserId int
	Role   string
}

type TeamUserLinkDTO struct {
	TeamId int    `param:"id" validate:"required"`
	UserId int    `json:"userId" validate:"required"`
	Role   string `json:"role" validate:"required"`
}

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
