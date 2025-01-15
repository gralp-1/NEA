package main

import (
	"image"
	"os"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SaveLoadWindow struct {
	Showing        bool
	Anchor         rl.Vector2
	InteractedWith time.Time
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	image, _, err := image.Decode(f)
	return image, err
}
func (s *SaveLoadWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(s.Anchor.X, s.Anchor.Y, 400, 300)
}
func (s *SaveLoadWindow) Draw() {
	s.Showing = !gui.WindowBox(s.getRect(), Translate("window.save.title"))
	dragDropRect := rl.NewRectangle((s.Anchor.X)+10, (s.Anchor.Y)+30, (s.getRect().Width)-20, (s.getRect().Height)-80)
	rl.DrawRectangleLinesEx(dragDropRect, 2, rl.Red)
	if gui.Button(rl.NewRectangle(dragDropRect.X, dragDropRect.Y+dragDropRect.Height+10, dragDropRect.Width, 30), "Save file as "+string(state.Config.FileFormat)) {
		state.SaveImage()
	}
	// This is a way to make it centred without doing any measurement
	stashAlignment := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)

	gui.Label(dragDropRect, "Drag & Drop new files here")

	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, stashAlignment)
	if rl.IsFileDropped() && rl.CheckCollisionPointRec(rl.GetMousePosition(), dragDropRect) {
		InfoLog("New file dropped in, saving original")
		state.SaveImage()
		DebugLog("Image saved")
		paths := rl.LoadDroppedFiles() // only take the first one`
		if len(paths) > 1 {
			ErrorLogf("More than one file dropped, only using the first one (%v)", paths[0])
		}
		InfoLogf("Loading image %v", paths[0])
		state.LoadImage(paths[0])
		DebugLog("Image loaded")
		state.RefreshImage()
		DebugLog("Image refreshed")
	}
}
