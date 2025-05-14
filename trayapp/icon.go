package trayapp

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"runtime"
	"sync/atomic"
	"time"

	"fyne.io/systray"
	"github.com/ParetoSecurity/agent/shared"
	"github.com/caarlos0/log"
	"github.com/fyne-io/image/ico"
)

// SetTemplateIcon sets the system tray icon based on the operating system.
type IconBadge string

const (
	BadgeNone    IconBadge = "none"
	BadgeOrange  IconBadge = "orange"
	BadgeGreen   IconBadge = "green"
	BadgeRunning IconBadge = "running" // New badge type for running state
)

// blinkCancelChan is used to signal when to stop blinking
var blinkCancelChan = make(chan struct{})

// isBlinking tracks if the icon is currently blinking
var isBlinking atomic.Bool

// setIcon sets the system tray icon based on the OS and theme.
func setIcon() {
	state := BadgeNone
	if !shared.GetModifiedTime().IsZero() {
		if shared.AllChecksPassed() {
			state = BadgeGreen
		} else {
			state = BadgeOrange
		}
	}
	if runtime.GOOS == "windows" {
		// Try to detect Windows theme (light/dark) and set icon accordingly
		icon := shared.IconBlack // fallback
		if IsDarkTheme() {
			icon = shared.IconWhite
		}
		SetTemplateIcon(icon, state)
		return
	}
	log.Debug("Setting icon for non-Windows OS")
	SetTemplateIcon(shared.IconWhite, state)
}

// startBlinkingIcon starts blinking the icon to indicate checks are running
// It will continue blinking until stopBlinkingIcon is called
func startBlinkingIcon() {
	// If already blinking, don't start another goroutine
	if isBlinking.Load() {
		return
	}
	isBlinking.Store(true)

	// Create a new channel first
	newCancelChan := make(chan struct{})

	// Store the old channel for closing
	oldCancelChan := blinkCancelChan

	// Update the global variable
	blinkCancelChan = newCancelChan

	// Close the old channel if it exists
	if oldCancelChan != nil {
		close(oldCancelChan)
	}

	go func(cancelCh chan struct{}) {
		blinkTicker := time.NewTicker(300 * time.Millisecond)
		defer blinkTicker.Stop()

		showBadge := true

		for {
			select {
			case <-cancelCh:
				// When canceled, make sure we update the icon to the current state
				setIcon()
				return
			case <-blinkTicker.C:
				// Toggle between showing the running badge and no badge
				if runtime.GOOS == "windows" {
					icon := shared.IconBlack
					if IsDarkTheme() {
						icon = shared.IconWhite
					}

					if showBadge {
						SetTemplateIcon(icon, BadgeRunning)
					} else {
						SetTemplateIcon(icon, BadgeNone)
					}
				} else {
					if showBadge {
						SetTemplateIcon(shared.IconWhite, BadgeRunning)
					} else {
						SetTemplateIcon(shared.IconWhite, BadgeNone)
					}
				}
				showBadge = !showBadge
			}
		}
	}(blinkCancelChan)
}

// stopBlinkingIcon stops the blinking effect and reverts to the normal icon
func stopBlinkingIcon() {
	if !isBlinking.Load() {
		return
	}
	isBlinking.Store(false)

	// Signal blinking to stop
	if blinkCancelChan != nil {
		close(blinkCancelChan)
		blinkCancelChan = nil
	}

	// Restore the normal icon
	setIcon()
}

// renderBadge overlays a colored dot (badge) onto the icon PNG bytes.
// Only supports orange and green for now.
func renderBadge(icon []byte, badge IconBadge) []byte {
	if badge == BadgeNone {
		return icon
	}
	img, err := png.Decode(bytes.NewReader(icon))
	if err != nil {
		log.WithError(err).Error("failed to decode PNG for badge rendering")
		return icon
	}
	bounds := img.Bounds()
	// Draw a small circle in the bottom right corner
	dotSize := bounds.Dx() / 2
	centerX := bounds.Max.X - dotSize
	centerY := bounds.Max.Y - dotSize

	// Create a new RGBA image to draw on
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{}, draw.Src)

	var dotColor color.Color
	switch badge {
	case BadgeOrange:
		dotColor = color.RGBA{R: 255, G: 140, B: 0, A: 255} // orange
	case BadgeGreen:
		dotColor = color.RGBA{R: 0, G: 200, B: 0, A: 255} // green
	case BadgeRunning:
		dotColor = color.RGBA{R: 255, G: 165, B: 0, A: 255} // bright orange
	default:
		return icon
	}

	// Draw the dot
	for y := 0; y < dotSize; y++ {
		for x := 0; x < dotSize; x++ {
			dx := x - dotSize/2
			dy := y - dotSize/2
			if dx*dx+dy*dy <= (dotSize/2)*(dotSize/2) {
				rgba.Set(centerX+x, centerY+y, dotColor)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		log.WithError(err).Error("failed to encode PNG with badge")
		return icon
	}
	return buf.Bytes()
}

func SetTemplateIcon(icon []byte, badge IconBadge) {
	iconWithBadge := renderBadge(icon, badge)
	if runtime.GOOS == "windows" {
		var icoBuffer bytes.Buffer
		pngImage, err := png.Decode(bytes.NewReader(iconWithBadge))
		if err != nil {
			log.WithError(err).Error("failed to decode PNG image")
		}
		if err := ico.Encode(&icoBuffer, pngImage); err != nil {
			log.WithError(err).Error("failed to encode ICO image")
		}
		systray.SetTemplateIcon(icoBuffer.Bytes(), icoBuffer.Bytes())
		return
	}
	log.Info("Setting icon for non-Windows OS")
	systray.SetTemplateIcon(iconWithBadge, iconWithBadge)
}
