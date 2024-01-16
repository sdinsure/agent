package swagger

import (
	"embed"
	"net/http"
)

//go:embed swaggerui/*
var SwaggerUIAssets embed.FS
var SwaggerUIAssetsHttpHandler = http.FileServer(http.FS(SwaggerUIAssets))
