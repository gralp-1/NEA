package main

import (
	"fmt"
	"math"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// global state variable
var state State

// TODO: use go image api and late stage convert to raylib image for easy image manipulation
func main() {
	rl.InitWindow(0, 0, "Image editor")
	defer rl.CloseWindow() // this makes sure that the window is always closed at the end of the function
	rl.SetTargetFPS(60)

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

		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}
}
