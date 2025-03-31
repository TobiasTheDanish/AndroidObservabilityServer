package model

type NewOwnerData struct {
	Name string
}

type NewApiKeyData struct {
	Key     string
	OwnerID int
}
