package main

import (
	"testing"
)

//func TestState(t *testing.T) {
//	t.Run("Test initialisation of state", func(t *testing.T) {
//		// initialise the state
//		state := State{}
//		state.Init()
//
//		// check that the shown image is the same as the original image
//		if state.ShownImage.ToImage() != state.OrigImage.ToImage() {
//			t.Errorf("Expected shown image to be the same as the original image, but got %v and %v", state.ShownImage, state.OrigImage)
//		}
//
//		// check that the background colour is white
//		if state.BackgroundColour != rl.RayWhite {
//			t.Errorf("Expected background colour to be white, but got %v", state.BackgroundColour)
//		}
//	})
//}

func BenchmarkState_ApplyFilters(b *testing.B) {
	// initialise the state
	state := State{}
	state.Init()
	state.IsGrayscaleEnabled = true
	state.IsDitheringEnabled = true
	b.StartTimer()

	// run the ApplyFilters function
	for i := 0; i < b.N; i++ {
		state.ApplyFilters()
	}
}
