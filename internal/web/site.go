package web

import (
	"io/fs"
	"net/http"

	"embed"
)

//go:embed site/index.html site/style.css site/app.js
var assets embed.FS

func Page() http.Handler {
	root, err := fs.Sub(assets, "site")
	if err != nil {
		return http.NotFoundHandler()
	}
	return http.FileServer(http.FS(root))
}
