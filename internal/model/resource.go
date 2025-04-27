package model

type NewMemoryUsageData struct {
	Id                 string
	SessionId          string
	InstallationId     string
	FreeMemory         int64
	UsedMemory         int64
	MaxMemory          int64
	TotalMemory        int64
	AvailableHeapSpace int64
}

type MemoryUsageEntity struct {
	Id                 string
	SessionId          string
	InstallationId     string
	FreeMemory         int64
	UsedMemory         int64
	MaxMemory          int64
	TotalMemory        int64
	AvailableHeapSpace int64
}
