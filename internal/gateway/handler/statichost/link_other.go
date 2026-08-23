//go:build !windows

package statichost

import "os"

func isWindowsReparsePoint(os.FileInfo) bool {
	return false
}
