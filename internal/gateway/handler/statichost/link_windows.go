//go:build windows

package statichost

import (
	"os"
	"syscall"
)

func isWindowsReparsePoint(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok || data == nil {
		return false
	}
	return data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
