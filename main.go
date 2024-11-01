package main

import (
	"fmt"
	"image/png"
	"math"
	"os"
	"strings"

	"github.com/cnf/structhash"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// global state variable
var state State

func main() {
	rl.InitWindow(0, 0, "Image editor")
	defer rl.CloseWindow() // this makes sure that the window is always closed at the end of the function
	rl.SetTargetFPS(60)

	state.Init()

	// adjust the window size so it's the same size as the image + a 400 pixel gutter for image controls
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))

	oldFiltersHash, _ := structhash.Hash(state.Filters, 1)

	state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(state.BackgroundColour)

		// apply all the filters
		// only reload the texture if the filters have changed because it's quite slow
		oldFiltersHash, _ = structhash.Hash(state.Filters, 1)
		rl.DrawTexture(state.CurrentTexture, 0, 0, rl.White)

		if rl.IsKeyPressed(rl.KeyG) {
			state.GenerateNoiseImage(500, 500)
		}
		// DRAW UI
		state.Filters.IsGrayscaleEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 10, 10, 10),
			"Grayscale",
			state.Filters.IsGrayscaleEnabled,
		)
		state.Filters.IsDitheringEnabled = gui.CheckBox( // gabagool
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 30, 10, 10),
			"Dithering",
			state.Filters.IsDitheringEnabled,
		)
		state.Filters.IsQuantizingEnabled = gui.CheckBox( // gabagool
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 50, 10, 10),
			"Quantization",
			state.Filters.IsQuantizingEnabled,
		)
		state.Filters.QuantizingBands = uint8(math.Trunc(float64(gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 70, 100, 10),
			fmt.Sprintf("Quantization bands: %d   3", state.Filters.QuantizingBands),
			"255",
			float32(state.Filters.QuantizingBands),
			3.0,
			16.0,
		))))
		state.Filters.ChannelAdjustmentEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 100, 10, 10),
			"Channel adjustment",
			state.Filters.ChannelAdjustmentEnabled,
		)
		state.Filters.ChannelAdjustment[0] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 110, 100, 10),
			"Red 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[0],
			0.0,
			1.0,
		)

		state.Filters.ChannelAdjustment[1] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 120, 100, 10),
			"Green 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[1],
			0.0,
			1.0,
		)
		state.Filters.ChannelAdjustment[2] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 130, 100, 10),
			"Blue 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[2],
			0.0,
			1.0,
		)
		if rl.IsKeyPressed(rl.KeyC) {
			// TODO: offset the window so none of it ever draws outside the window
			state.FilterWindow.Anchor = rl.GetMousePosition()
			state.FilterWindow.Showing = !state.FilterWindow.Showing
		}
		if state.FilterWindow.Showing {
			state.FilterWindow.Draw()
		}

		newFiltersHash, _ := structhash.Hash(state.Filters, 1)
		if strings.Compare(newFiltersHash, oldFiltersHash) != 0 {
			state.RefreshImage()
		}

		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}

	f, _ := os.Create("image.png")
	_ = png.Encode(f, state.WorkingImage.SubImage(state.WorkingImage.Rect)) // what a stupid API
}
