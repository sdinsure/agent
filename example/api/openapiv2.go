package api

import (
	"embed"
	"net/http"
)

//go:embed openapiv2/*
var OpenApiV2Fs embed.FS
var OpenApiV2HttpHandler = http.FileServer(http.FS(OpenApiV2Fs))
