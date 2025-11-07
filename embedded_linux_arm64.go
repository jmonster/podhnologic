//go:build linux && arm64

package main

import "embed"

//go:embed binaries/linux-arm64/*
var embeddedBinariesFS embed.FS
