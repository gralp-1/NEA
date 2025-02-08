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
	p.Showing = !gui.WindowBox(p.getRect(), Translate("window.help.title"))
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+30, 300, 40), "H - "+Translate("window.help.help"))
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+50, 300, 40), "C - "+Translate("window.help.palette"))
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+70, 300, 40), "O - "+Translate("window.help.order"))
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+90, 300, 40), "S - "+Translate("window.help.save"))
	gui.Label(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+110, 300, 40), ", - "+Translate("window.help.settings"))
}
