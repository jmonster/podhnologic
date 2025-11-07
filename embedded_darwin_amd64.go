//go:build darwin && amd64

package main

import "embed"

//go:embed binaries/darwin-amd64/*
var embeddedBinariesFS embed.FS
