package main

import "golang.org/x/sys/windows"

func publishOutput(tempPath, outputPath string) error {
	from, err := windows.UTF16PtrFromString(tempPath)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(outputPath)
	if err != nil {
		return err
	}
	// With no MOVEFILE_REPLACE_EXISTING flag, an existing destination fails.
	return windows.MoveFileEx(from, to, windows.MOVEFILE_WRITE_THROUGH)
}
