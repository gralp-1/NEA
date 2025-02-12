package main

import (
	"fmt"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SettingsWindow struct {
	Showing                  bool
	Anchor                   rl.Vector2
	InteractedWith           time.Time
	ActiveLanguage           Language
	IsLanguageDropDownActive bool
	IsThemeDropDownActive    bool
	IsFontDropDownActive     bool
}

func (w *SettingsWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(w.Anchor.X, w.Anchor.Y, 400, 300)
}
func (w *SettingsWindow) Draw() {
	w.Showing = !gui.WindowBox(w.getRect(), Translate("window.settings.title"))
	
	// Draw the font size label always in font size 10 so it's readable
	gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, 10)
	gui.Label(rl.NewRectangle(w.Anchor.X+70, w.Anchor.Y+70, 100, 7), fmt.Sprintf("Font Size: %d", state.Config.FontSize))
	state.SetFontSize()

	storeFontSize := state.Config.FontSize
	// font size slider
	state.Config.FontSize = int64(gui.Slider(rl.NewRectangle(w.Anchor.X+70, w.Anchor.Y+90, 100, 7), "10", "42", float32(state.Config.FontSize), 10.0, 42.0))
	
	// set the font size if it changes
	if state.Config.FontSize != storeFontSize {
		state.SetFontSize()
	}

	// Language selection
	if gui.DropdownBox(rl.NewRectangle(w.Anchor.X+10, w.Anchor.Y+30, 100, 30), "English;Deutsch", (*int32)(&state.Config.Language), w.IsLanguageDropDownActive) {
		w.IsLanguageDropDownActive = !w.IsLanguageDropDownActive
		state.LoadLanguageData()
	}
	
	// theme selection
	if gui.DropdownBox(rl.NewRectangle(w.Anchor.X+120, w.Anchor.Y+30, 100, 30), "Light;Dark", (*int32)(&state.Config.CurrentTheme), w.IsThemeDropDownActive) {
		w.IsThemeDropDownActive = !w.IsThemeDropDownActive
		state.ChangeTheme()
	}
	storedFont := state.Config.CurrentFont
	// font selector
	if gui.DropdownBox(rl.NewRectangle(w.Anchor.X+230, w.Anchor.Y+30, 100, 30), "Default;Arial;Berkley Mono;Comic Sans;Zapfino;Spleen", (*int32)(&state.Config.CurrentFont), w.IsFontDropDownActive) {
		w.IsFontDropDownActive = !w.IsFontDropDownActive
		if state.Config.CurrentFont != storedFont {
			state.ChangeFont()
		}
	}
}
