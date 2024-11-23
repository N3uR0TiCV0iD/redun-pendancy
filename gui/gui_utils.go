package gui

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func BoldTextStyle() fyne.TextStyle {
	return fyne.TextStyle{Bold: true}
}

func DefaultForegroundColor() color.Color {
	return theme.Color(theme.ColorNameForeground)
}

func DefaultBackgroundColor() color.Color {
	return theme.Color(theme.ColorNameBackground)
}

func DisabledColor() color.Color {
	return theme.Color(theme.ColorNameDisabled)
}

func FocusWidget(window fyne.Window, widget fyne.Focusable) {
	window.Canvas().Focus(widget)
}

func TrySetWindowIcon(window fyne.Window, iconPath string) {
	icon, err := fyne.LoadResourceFromPath(iconPath)
	if err != nil {
		log.Printf("[Warning] Failed to load window icon: %v", err)
		return
	}
	window.SetIcon(icon)
}
