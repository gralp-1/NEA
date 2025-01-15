package main

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/cnf/structhash"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// TODO: config system
// global state variable
var state State
var globalQuantizeLUT [][]uint8

func main() {
	rl.InitWindow(0, 0, "")
	defer rl.CloseWindow() // this makes sure that the window is always closed at the end of the function
	rl.SetTargetFPS(60 * 2)
	InitLut()

	state.Init()

	// adjust the window size so it's the same size as the image + a 400 pixel gutter for image controls
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))

	oldFiltersHash, _ := structhash.Hash(state.Filters, 1)

	state.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage)
	state.HelpWindow.Showing = true
	state.HelpWindow.InteractedWith = time.Now()
	rl.SetWindowTitle(Translate("main.title"))
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(state.BackgroundColour)

		// apply all the filters
		// only reload the texture if the filters have changed because it's quite slow
		oldFiltersHash, _ = structhash.Hash(state.Filters, 1)
		rl.DrawTexture(state.CurrentTexture, 0, 0, rl.White)

		//if rl.IsKeyPressed(rl.KeyG) {
		//	state.GenerateNoiseImage(500, 500)
		//}
		// DRAW UI
		state.Filters.IsGrayscaleEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 10, 10, 10),
			Translate("control.grayscale"),
			state.Filters.IsGrayscaleEnabled,
		)
		state.Filters.IsDitheringEnabled = gui.CheckBox( // gabagool
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 40, 10, 10),
			Translate("control.dithering"),
			state.Filters.IsDitheringEnabled,
		)
		state.Filters.DitheringQuantizationBuckets = uint8(math.Trunc(float64(gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 55, 100, 10),
			"0", "15", float32(state.Filters.DitheringQuantizationBuckets), 0.0, 15.0))))

		state.Filters.IsQuantizingEnabled = gui.CheckBox( // gabagool
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 85, 10, 10),
			Translate("control.quantizing"),
			state.Filters.IsQuantizingEnabled,
		)
		state.Filters.QuantizingBands = uint8(math.Trunc(float64(gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 100, 100, 10),
			fmt.Sprintf("%s: %d   3", Translate("control.quantizationbands"), state.Filters.QuantizingBands),
			"255",
			float32(state.Filters.QuantizingBands),
			3.0,
			16.0,
		))))
		state.Filters.ChannelAdjustmentEnabled = gui.CheckBox(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 130, 10, 10),
			Translate("control.channeladjustment"),
			state.Filters.ChannelAdjustmentEnabled,
		)
		state.Filters.ChannelAdjustment[0] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 145, 100, 10),
			fmt.Sprintf("%s 0.0", Translate("colour.red")),
			"1.0",
			state.Filters.ChannelAdjustment[0],
			0.0,
			1.0,
		)

		state.Filters.ChannelAdjustment[1] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 160, 100, 10),
			fmt.Sprintf("%s 0.0", Translate("colour.green")),
			"1.0",
			state.Filters.ChannelAdjustment[1],
			0.0,
			1.0,
		)
		state.Filters.ChannelAdjustment[2] = gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 175, 100, 10),
			fmt.Sprintf("%s 0.0", Translate("colour.blue")),
			"1.0",
			state.Filters.ChannelAdjustment[2],
			0.0,
			1.0,
		)

		state.Filters.IsBoxBlurEnabled = gui.CheckBox(rl.NewRectangle(float32(rl.GetScreenWidth()-200), 205, 10, 10), Translate("control.boxblur"), state.Filters.IsBoxBlurEnabled)
		// gaussian deviation
		state.Filters.BoxBlurIterations = int(gui.Slider(
			rl.NewRectangle(float32(rl.GetScreenWidth()-200), 225, 100, 10),
			fmt.Sprintf("%s: %d 1 ", Translate("control.boxblur.iterations"), state.Filters.BoxBlurIterations),
			"10",
			float32(state.Filters.BoxBlurIterations),
			1.0,
			10.0,
		))

		mousePos := rl.GetMousePosition()
		if rl.IsKeyPressed(rl.KeyO) {
			DebugLog("Toggling filter order window")
			state.FilterWindow.Anchor = rl.Vector2{
				X: min(mousePos.X, float32(rl.GetScreenWidth()-int(state.FilterWindow.getRect().Width))),
				Y: min(mousePos.Y, float32(rl.GetScreenHeight()-int(state.FilterWindow.getRect().Height))),
			}
			state.FilterWindow.Showing = !state.FilterWindow.Showing
			state.FilterWindow.InteractedWith = time.Now()
		}
		if rl.IsKeyPressed(rl.KeyC) {
			DebugLog("Toggling palette window")
			state.PaletteWindow.Anchor = rl.Vector2{
				X: min(mousePos.X, float32(rl.GetScreenWidth()-int(state.PaletteWindow.getRect().Width))),
				Y: min(mousePos.Y, float32(rl.GetScreenHeight()-int(state.PaletteWindow.getRect().Height))),
			}
			state.PaletteWindow.Showing = !state.PaletteWindow.Showing
			state.PaletteWindow.InteractedWith = time.Now()
		}
		if rl.IsKeyPressed(rl.KeyH) {
			DebugLog("Toggling help window")
			state.HelpWindow.Anchor = rl.Vector2{
				X: min(mousePos.X, float32(rl.GetScreenWidth()-int(state.HelpWindow.getRect().Width))),
				Y: min(mousePos.Y, float32(rl.GetScreenHeight()-int(state.HelpWindow.getRect().Height))),
			}
			state.HelpWindow.Showing = !state.HelpWindow.Showing
			state.HelpWindow.InteractedWith = time.Now()
		}
		if rl.IsKeyPressed(rl.KeyS) {
			DebugLog("Toggling save & load window")
			state.SaveLoadWindow.Anchor = rl.Vector2{
				X: min(mousePos.X, float32(rl.GetScreenWidth()-int(state.SaveLoadWindow.getRect().Width))),
				Y: min(mousePos.Y, float32(rl.GetScreenHeight()-int(state.SaveLoadWindow.getRect().Height))),
			}
			state.SaveLoadWindow.Showing = !state.SaveLoadWindow.Showing
			state.SaveLoadWindow.InteractedWith = time.Now()
		}
		if rl.IsKeyPressed(rl.KeyComma) {
			DebugLog("Toggling settings window")
			state.SettingsWindow.Anchor = rl.Vector2{
				X: min(mousePos.X, float32(rl.GetScreenWidth()-int(state.SettingsWindow.getRect().Width))),
				Y: min(mousePos.Y, float32(rl.GetScreenHeight()-int(state.SettingsWindow.getRect().Height))),
			}
			state.SettingsWindow.Showing = !state.SettingsWindow.Showing
			state.SettingsWindow.InteractedWith = time.Now()
		}
		if rl.IsKeyPressed(rl.KeyQ) {
			state.Close()
		}

		times := []int64{state.HelpWindow.InteractedWith.Unix(), state.PaletteWindow.InteractedWith.Unix(), state.FilterWindow.InteractedWith.Unix(), state.SaveLoadWindow.InteractedWith.Unix(), state.SettingsWindow.InteractedWith.Unix()}
		// make a list of the windows sorted by newest interacted with to new
		// NOTE: need to modify this on new window
		slices.Sort(times)
		for _, t := range times {
			switch t {
			case state.FilterWindow.InteractedWith.Unix():
				if state.FilterWindow.Showing {
					state.FilterWindow.Draw()
				}
			case state.HelpWindow.InteractedWith.Unix():
				if state.HelpWindow.Showing {
					state.HelpWindow.Draw()
				}
			case state.PaletteWindow.InteractedWith.Unix():
				if state.PaletteWindow.Showing {
					state.PaletteWindow.Draw()
				}
			case state.SaveLoadWindow.InteractedWith.Unix():
				if state.SaveLoadWindow.Showing {
					state.SaveLoadWindow.Draw()
				}
			case state.SettingsWindow.InteractedWith.Unix():
				if state.SettingsWindow.Showing {
					state.SettingsWindow.Draw()
				}
			}
		}

		newFiltersHash, _ := structhash.Hash(state.Filters, 1)
		if strings.Compare(newFiltersHash, oldFiltersHash) != 0 {
			state.RefreshImage()
		}

		rl.DrawFPS(10, 10)
		rl.EndDrawing()
	}
	state.Close()
}

// BUG: QuantizeValue when BucketCount = 8 and 255
