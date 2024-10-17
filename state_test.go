package main

import (
	"image"
	"slices"
	"testing"
)

func TestGrayscale(t *testing.T) {
	t.Run("RGB Channels", func(t *testing.T) {
		// Aim: Grayscale filter should average the RGB values of each pixel

		s := State{
			WorkingImage: image.RGBA{
				Pix: []uint8{255, 0, 0, 255, 0, 99, 0, 255}, // Red, Green
			},
		}
		s.GrayscaleFilter()
		if s.WorkingImage.Pix[0] != 85 || s.WorkingImage.Pix[1] != 85 || s.WorkingImage.Pix[2] != 85 {
			t.Error("Grayscale filter failed, expected 85, 85, 85, got", s.WorkingImage.Pix[0], s.WorkingImage.Pix[1], s.WorkingImage.Pix[2])
		}
		if s.WorkingImage.Pix[4] != 33 || s.WorkingImage.Pix[5] != 33 || s.WorkingImage.Pix[6] != 33 {
			t.Error("Grayscale filter failed, expected 33, 33, 33, got", s.WorkingImage.Pix[4], s.WorkingImage.Pix[5], s.WorkingImage.Pix[6])
		}
	})
	t.Run("Alpha values", func(t *testing.T) {
		// Aim: Grayscale filter should not mutate alpha values
		s := State{
			WorkingImage: image.RGBA{
				Pix: []uint8{255, 0, 0, 34, 0, 255, 0, 75}, // Red, Green, Blue pixels
			},
		}
		s.GrayscaleFilter()
		if s.WorkingImage.Pix[3] != 34 || s.WorkingImage.Pix[7] != 75 {
			t.Error("Grayscale filter mutated alpha values: ", s.WorkingImage.Pix)
		}
	})
}

func TestQuantization(t *testing.T) {
	t.Run("Simple Quantization", func(t *testing.T) {
		s := State{
			WorkingImage: image.RGBA{
				Pix: []uint8{255, 0, 0, 255, 0, 99, 0, 255}, // Red, Green
			},
		}
		s.QuantizingFilter()

	})
}

func TestRemoveDuplicates(t *testing.T) {
	t.Run("No duplicates", func(t *testing.T) {
		input := []int{1, 2, 3, 4, 5}
		if slices.Compare(removeDuplicates(input), input) != 0 {
			t.Fail()
		}
	})
	t.Run("Some duplicates", func(t *testing.T) {
		input := []int{1, 1, 2, 3, 4}
		if slices.Compare(removeDuplicates(input), []int{1, 2, 3, 4}) != 0 {
			t.Fail()
		}
	})
	t.Run("All duplicates", func(t *testing.T) {
		input := []int{1, 1, 1, 1, 1}
		if slices.Compare(removeDuplicates(input), []int{1}) != 0 {
			t.Fail()
		}
	})
}

func TestImagePallette(t *testing.T) {
	t.Run("Palletisation", func(t *testing.T) {
		s := State{
			WorkingImage: image.RGBA{
				Pix: []uint8{255, 0, 0, 255, 0, 99, 0, 255}, // Red, Green
			},
		}
		s.GenerateImagePalette()
	})
	// t.Run("Multiple colour reduction", f func(t *testing.T))
}
