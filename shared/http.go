package shared

import (
	"fmt"
	"runtime"
)

func UserAgent() string {
	platform := runtime.GOOS
	if runtime.GOOS == "darwin" {
		platform = "macos"
	}
	return fmt.Sprintf("ParetoSecurity/agent %s/%s", platform, Version)
}
