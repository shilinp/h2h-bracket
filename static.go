package starter

import "embed"

//go:embed web/dist/*
var StaticFiles embed.FS
