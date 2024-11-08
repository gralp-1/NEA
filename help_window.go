package main

import (
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type HelpWindow struct {
	Showing        bool
	Anchor         rl.Vector2
	InteractedWith time.Time
}

func (p *HelpWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 500, 375)
}
func (p *HelpWindow) Draw() {
	p.Showing = !gui.WindowBox(p.getRect(), "Help window")
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+30, 2_00, 30), "H - Open this help window")
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+50, 2_00, 30), "C - View palette")
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+70, 2_00, 30), "O - Change filter order window")
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+90, 2_00, 30), "S - Open save window")
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+110, 200, 30), ", - Open settings window")
}
