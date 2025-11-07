//go:build linux && amd64

package main

import "embed"

//go:embed binaries/linux-amd64/*
var embeddedBinariesFS embed.FS
