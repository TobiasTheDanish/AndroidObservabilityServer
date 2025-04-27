package model

type NewMemoryUsageDTO struct {
	Id                 string `json:"id" validate:"required,uuid"`
	SessionId          string `json:"sessionId" validate:"required,uuid"`
	InstallationId     string `json:"installationId" validate:"required,uuid"`
	FreeMemory         int64  `json:"freeMemory"`
	UsedMemory         int64  `json:"usedMemory"`
	MaxMemory          int64  `json:"maxMemory"`
	TotalMemory        int64  `json:"totalMemory"`
	AvailableHeapSpace int64  `json:"availableHeapSpace"`
}

type NewMemoryUsageData struct {
	Id                 string
	SessionId          string
	InstallationId     string
	AppId              int
	FreeMemory         int64
	UsedMemory         int64
	MaxMemory          int64
	TotalMemory        int64
	AvailableHeapSpace int64
}

type GetMemoryUsageDTO struct {
	Id                 string `json:"id"`
	SessionId          string `json:"sessionId"`
	InstallationId     string `json:"installationId"`
	AppId              int    `json:"appId"`
	FreeMemory         int64  `json:"freeMemory"`
	UsedMemory         int64  `json:"usedMemory"`
	MaxMemory          int64  `json:"maxMemory"`
	TotalMemory        int64  `json:"totalMemory"`
	AvailableHeapSpace int64  `json:"availableHeapSpace"`
}

type MemoryUsageEntity struct {
	Id                 string
	SessionId          string
	InstallationId     string
	AppId              int
	FreeMemory         int64
	UsedMemory         int64
	MaxMemory          int64
	TotalMemory        int64
	AvailableHeapSpace int64
}
