//go:build unix
// +build unix

package shared

import (
	_ "embed"
)

var (
	//go:embed icon_white.png
	IconWhite []byte

	//go:embed icon_black.png
	IconBlack []byte

	//go:embed icon_large.png
	IconLarge []byte
)
