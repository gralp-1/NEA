package main

import (
	"math"
	"slices"
	"strings"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type PaletteWindow struct {
	Showing         bool
	Anchor          rl.Vector2
	InteractedWith  time.Time
	ActiveHistogram int32
}

func (p *PaletteWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 796, 626)
	// return rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 596, 426)
}
func (p *PaletteWindow) DrawHistogram(anchor rl.Vector2, data []int, colour rl.Color) {
	data = data[1:]
	largestFreq := slices.Max(data)
	if largestFreq == 0 {
		ErrorLog("Largest histogram value was zero")
		gui.Label(rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 100, 20), Translate("window.palette.emptyerror"))
		return
	}

	// Draw title
	rl.DrawRectangle(int32(anchor.X)+5, int32(anchor.Y)+5, 20, 20, colour)
	rl.DrawText(Translate("window.palette.title"), int32(anchor.X)+30, int32(anchor.Y)+6, 20, rl.Black)
	
	// get metrics for drawing the graph
	var GraphWidth = int32(p.getRect().Width - 20.0)
	var BarWidth = int32(math.Floor(float64(GraphWidth / 256.0)))
	var GraphHeight = int32(p.getRect().Height - 50.0)


	// for each value in the channel
	for i := int32(0); i < 255; i++ {
		// get the values relative occurance
		fractionalHeight := float64(data[i]) / float64(largestFreq) // [0, 1]
		// multiply that by the max height of a bar
		height := int32(float64(GraphHeight) * fractionalHeight)
		// get how far across to draw it
		x := int32(anchor.X + float32(i*BarWidth))
		// draw the rectangle
		rl.DrawRectangle(x, int32(anchor.Y)+GraphHeight-height+20, BarWidth, height-25, colour)
	}
}
func (p *PaletteWindow) Draw() {
	p.Showing = !gui.WindowBox(p.getRect(), Translate("window.palette.title"))
	
	// select the channel to draw
	controlString := strings.Join(MapOut([]string{"colour.red", "colour.blue", "colour.green"}, Translate), ";")
	p.ActiveHistogram = gui.ToggleGroup(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+30, 100, 20), controlString, p.ActiveHistogram)

	// draw the histogram
	switch p.ActiveHistogram {
	case 0:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.RedHistogram[:], rl.Red)
	case 1:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.GreenHistogram[:], rl.Green)
	case 2:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.BlueHistogram[:], rl.Blue)
	}
}
