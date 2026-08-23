package statichost

import "os"

// isLinkFile 判断路径是否为符号链接或 Windows 连接点。
func isLinkFile(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	return isWindowsReparsePoint(info)
}
