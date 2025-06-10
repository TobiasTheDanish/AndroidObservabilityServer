package doc

import (
	"embed"
	"net/http"
)

//go:embed assets index.html
var content embed.FS

func DocFiles() http.FileSystem {
	return http.FS(content)
}
