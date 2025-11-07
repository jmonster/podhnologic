//go:build darwin && arm64

package main

import "embed"

//go:embed binaries/darwin-arm64/*
var embeddedBinariesFS embed.FS
