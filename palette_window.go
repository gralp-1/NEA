package main

import (
	"math"
	"slices"
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

func getBrightness(colour rl.Color) float64 {
	return 0.2126*float64(colour.R) + 0.7152*float64(colour.G) + 0.0722*float64(colour.B)
}
func (p *PaletteWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 796, 626)
	// return rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 596, 426)
}
func (p *PaletteWindow) DrawHistogram(anchor rl.Vector2, data []int, colour rl.Color) {
	// TODO: add axis
	// TODO: overlay all at once with semi transparent fill
	// TODO: line mode
	// TODO: highlight hovered line and display tooltip with frequency
	// TODO: make the histogram not just draw thin lines when quantized
	data = data[1:]
	largestFreq := slices.Max(data)
	if largestFreq == 0 {
		ErrorLog("Largest histogram value was zero")
		gui.Label(rl.NewRectangle(p.Anchor.X, p.Anchor.Y, 100, 20), "Image contains none of this colour")
		return
	}

	// Draw title
	rl.DrawRectangle(int32(anchor.X)+5, int32(anchor.Y)+5, 20, 20, colour)
	rl.DrawText("Channel Histogram", int32(anchor.X)+30, int32(anchor.Y)+6, 20, rl.Black)

	var GraphWidth = int32(p.getRect().Width - 20.0)
	var BarWidth = int32(math.Floor(float64(GraphWidth / 256.0)))
	var GraphHeight = int32(p.getRect().Height - 50.0)

	for i := int32(0); i < 255; i++ {
		fractionalHeight := float64(data[i]) / float64(largestFreq) // [0, 1]
		height := int32(float64(GraphHeight) * fractionalHeight)
		x := int32(anchor.X + float32(i*BarWidth))
		rl.DrawRectangle(x, int32(anchor.Y)+GraphHeight-height+20, BarWidth, height-25, colour)

	}
}
func (p *PaletteWindow) Draw() {
	p.Showing = !gui.WindowBox(p.getRect(), Translate("window.palette.title"))

	p.ActiveHistogram = gui.ToggleGroup(rl.NewRectangle(p.Anchor.X+10, p.Anchor.Y+30, 100, 20), "Red;Green;Blue", p.ActiveHistogram)

	switch p.ActiveHistogram {
	case 0:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.RedHistogram[:], rl.Red)
	case 1:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.GreenHistogram[:], rl.Green)
	case 2:
		p.DrawHistogram(rl.Vector2{X: p.Anchor.X + 5, Y: p.Anchor.Y + 50}, state.BlueHistogram[:], rl.Blue)
	}
}
