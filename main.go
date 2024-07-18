package main

import (
	"log"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// global state variable
var state State

// TODO: use go image api and late stage convert to raylib image for easy image manipulation
func main() {
	// initialise window
	rl.InitWindow(0, 0, "Image editor")
	defer rl.CloseWindow() // this makes sure that the window is always closed at the end of the function
	rl.SetTargetFPS(60)

	// initialise state
	state.Init()

	// adjust the window size so it's the same size as the image + a 400 pixel gutter for image controls
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))

	oldFilters := state.Filters
	state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(state.BackgroundColour)
		// apply all the filters

		if oldFilters != state.Filters {
			log.Println("Filters changed")
			state.ApplyFilters()
			state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)
		}
		oldFilters = state.Filters
		rl.DrawTexture(state.CurrentTexture, 0, 0, rl.White)

		// DRAW UI
		state.Filters.IsGrayscaleEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 10, 10, 10),
			"Grayscale",
			state.Filters.IsGrayscaleEnabled,
		)

		rl.DrawFPS(10, 10)

		rl.EndDrawing()
	}
}
