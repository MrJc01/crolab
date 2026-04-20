package web

import "embed"

//go:embed admin/* client/* lab/*
var StaticFiles embed.FS
