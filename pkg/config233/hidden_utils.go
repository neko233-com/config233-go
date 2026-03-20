package config233

import (
	"os"
	"strings"
)

func isHiddenDir(info os.FileInfo) bool {
	if info == nil || !info.IsDir() {
		return false
	}
	name := info.Name()
	return name != "." && name != ".." && strings.HasPrefix(name, ".")
}
