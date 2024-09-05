package main

import (
	"image"
	"testing"
)

func TestGrayscale(t *testing.T) {
	t.Run("Grayscale", func(t *testing.T) {
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
