package main

import "golang.org/x/sys/unix"

// publishOutput atomically moves completed audio without replacing another file.
func publishOutput(tempPath, outputPath string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, tempPath, unix.AT_FDCWD, outputPath, unix.RENAME_EXCL)
}
