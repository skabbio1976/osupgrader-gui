package gui

import "github.com/skabbio1976/fyne-autodpi"

// applyDPIScale delegates to fyne-autodpi to detect and set FYNE_SCALE
// before the Fyne application is created.
func applyDPIScale() {
	autodpi.MustApply()
}


