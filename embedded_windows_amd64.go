//go:build windows && amd64

package main

import "embed"

//go:embed binaries/windows-amd64/*
var embeddedBinariesFS embed.FS
