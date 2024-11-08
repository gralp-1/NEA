package main

import (
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"time"
)

type SettingsWindow struct {
	Showing                  bool
	Anchor                   rl.Vector2
	InteractedWith           time.Time
	ActiveLanguage           Language
	IsLanguageDropDownActive bool
}

func (w *SettingsWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(w.Anchor.X, w.Anchor.Y, 400, 300)
}
func (w *SettingsWindow) Draw() {
	w.Showing = !gui.WindowBox(w.getRect(), "Settings")

	// Language selection
	w.IsLanguageDropDownActive = gui.DropdownBox(rl.NewRectangle(w.Anchor.X+10, w.Anchor.Y+30, 100, 30), "English;German", (*int32)(&w.ActiveLanguage), w.IsLanguageDropDownActive)

	// Theme selection

	// Format selection
	// PNG, JPEG, TIFF
}
