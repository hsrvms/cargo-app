package web

import "embed"

//go:embed assets
var Assets embed.FS

//go:embed scripts
var Scripts embed.FS

//go:embed config
var Config embed.FS
