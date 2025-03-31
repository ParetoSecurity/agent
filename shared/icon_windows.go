//go:build windows
// +build windows

package shared

import (
	_ "embed"
)

var (
	//go:embed icon.ico
	IconWhite []byte

	//go:embed icon.ico
	IconBlack []byte

	//go:embed icon_large.png
	IconLarge []byte
)
