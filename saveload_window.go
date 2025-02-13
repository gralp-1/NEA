package main

import (
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SaveLoadWindow struct {
	Showing                  bool
	Anchor                   rl.Vector2
	InteractedWith           time.Time
	IsFileTypeDropDownActive bool
}

func (s *SaveLoadWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(s.Anchor.X, s.Anchor.Y, 400, 440)
}
func (s *SaveLoadWindow) Draw() {
	s.Showing = !gui.WindowBox(s.getRect(), Translate("window.save.title"))
	// TODO: filename text box
	// Save button
	if gui.Button(
		rl.NewRectangle(s.getRect().X+10, s.getRect().Y+30+5, s.getRect().Width-20, 30),
		"Save file") {
		state.SaveImage()
	}
	
	// File type dropdown
	if gui.DropdownBox(
		rl.NewRectangle(s.getRect().X+10, s.getRect().Y+30+45, s.getRect().Width-20, 30),
		"png;jpg;tiff;bmp",
		&state.Config.ActiveFormatIndex,
		s.IsFileTypeDropDownActive) {
		s.IsFileTypeDropDownActive = !s.IsFileTypeDropDownActive
	}

}
