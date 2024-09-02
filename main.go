package main

import (
	"fmt"
	"image/png"
	"math"
	"os"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// global state variable
var state State

func main() {
	rl.InitWindow(0, 0, "Image editor")
	defer rl.CloseWindow() // this makes sure that the window is always closed at the end of the function
	//rl.SetTargetFPS(60)

	state.Init()

	// adjust the window size so it's the same size as the image + a 400 pixel gutter for image controls
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))

	oldFilters := state.Filters
	state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(state.BackgroundColour)

		// apply all the filters
		// only reload the texture if the filters have changed because it's quite slow
		if oldFilters != state.Filters {
			InfoLog("Filters changed")
			state.ApplyFilters()
			// NOTE: Write about the performance and how it impacts UX and stuff
			state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage) // ~100ms
			// start := time.Now()
			// rl.UpdateTexture(state.CurrentTexture, Uint8SliceToRGBASlice(state.OrigImage.Pix))
			// log.Printf("Time to update texture: %s", time.Since(start).String())
		}
		oldFilters = state.Filters
		rl.DrawTexture(state.CurrentTexture, 0, 0, rl.White)

		// DRAW UI
		state.Filters.IsGrayscaleEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 10, 10, 10),
			"Grayscale",
			state.Filters.IsGrayscaleEnabled,
		)
		state.Filters.IsQuantizingEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 30, 10, 10),
			"Quantization",
			state.Filters.IsQuantizingEnabled,
		)
		state.Filters.QuantizingBands = uint8(math.Trunc(float64(gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 50, 100, 10),
			fmt.Sprintf("Quantization bands: %d   0", state.Filters.QuantizingBands),
			"255",
			float32(state.Filters.QuantizingBands),
			0.0,
			16.0,
		))))
		state.Filters.ChannelAdjustmentEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 80, 10, 10),
			"Channel adjustment",
			state.Filters.ChannelAdjustmentEnabled,
		)
		state.Filters.ChannelAdjustment[0] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 90, 100, 10),
			"Red 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[0],
			0.0,
			1.0,
		)

		state.Filters.ChannelAdjustment[1] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 100, 100, 10),
			"Green 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[1],
			0.0,
			1.0,
		)
		state.Filters.ChannelAdjustment[2] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 110, 100, 10),
			"Blue 0.0",
			"1.0",
			state.Filters.ChannelAdjustment[2],
			0.0,
			1.0,
		)

		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}
	f, _ := os.Create("image.png")
	_ = png.Encode(f, state.WorkingImage.SubImage(state.WorkingImage.Rect)) // what a stupid API
}
