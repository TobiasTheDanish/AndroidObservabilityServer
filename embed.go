package doc

import (
	"embed"
	"net/http"
)

//go:embed doc
var content embed.FS

func DocFiles() http.FileSystem {
	return http.FS(content)
}
