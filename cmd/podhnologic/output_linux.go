package main

import "golang.org/x/sys/unix"

func publishOutput(tempPath, outputPath string) error {
	return unix.Renameat2(unix.AT_FDCWD, tempPath, unix.AT_FDCWD, outputPath, unix.RENAME_NOREPLACE)
}
