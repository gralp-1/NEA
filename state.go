package main

import (
	"bytes"
	"image"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage image.RGBA
	ShownImage   *rl.Image

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
}

type Filters struct {
	IsGrayscaleEnabled  bool
	IsQuantizingEnabled bool
	QuantizingBands     uint8
	IsDitheringEnabled  bool
	DitheringLevel      int
}

func (s *State) GrayscaleFilter() {
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		mean := uint8((int(s.WorkingImage.Pix[i]) + int(s.WorkingImage.Pix[i+1]) + int(s.WorkingImage.Pix[i+2])) / 3)
		s.WorkingImage.Pix[i+0] = mean
		s.WorkingImage.Pix[i+1] = mean
		s.WorkingImage.Pix[i+2] = mean
	}
}
func (s *State) QuantizingFilter() {
	quantizationBandWidth := 255 / float64(state.Filters.QuantizingBands-1)
	// floor(x/bandWidth)*bandWidth + bandWidth/2
	// FIXME this is terrible, doesnn't work and crashes in weird edge cases
	res := make([]uint8, len(s.WorkingImage.Pix))
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		if i == 100 {
			DebugLogf("Pixel 1 orig: %d, %d, %d", s.WorkingImage.Pix[i+0], s.WorkingImage.Pix[i+1], s.WorkingImage.Pix[i+2])
		}
		s.WorkingImage.Pix[i+0] = uint8(math.Floor(float64(s.WorkingImage.Pix[i+0])/quantizationBandWidth)*quantizationBandWidth + (quantizationBandWidth / 2))
		s.WorkingImage.Pix[i+1] = uint8(math.Floor(float64(s.WorkingImage.Pix[i+1])/quantizationBandWidth)*quantizationBandWidth + (quantizationBandWidth / 2))
		s.WorkingImage.Pix[i+2] = uint8(math.Floor(float64(s.WorkingImage.Pix[i+2])/quantizationBandWidth)*quantizationBandWidth + (quantizationBandWidth / 2))
		s.WorkingImage.Pix[i+3] = s.OrigImage.Pix[i+3] // NOTE: forgot this line and image went invisible, write that down
		if i == 100 {
			DebugLogf("Pixel 1 post: %d, %d, %d", res[i+0], res[i+1], res[i+2])
		}
	}
}

func findClosestPalleteCol(col color.Color) color.Color {
	// quantize
	return nil
}
func subCol(lhs, rhs color.RGBA) color.RGBA {
	return color.RGBA{
		R: lhs.R - rhs.R,
		G: lhs.G - rhs.G,
		B: lhs.B - rhs.B,
		A: 255,
	}
}

//func (s *State) DitheringFilter() {
//	for y := range s.WorkingImage.Bounds().Dy() {
//		for x := range s.WorkingImage.Bounds().Dx() {
//			oldPixel := s.WorkingImage.At(x, y)
//			newPixel := findClosestPalleteCol(oldPixel)
//			idx := s.WorkingImage.PixOffset(x, y)
//			s.WorkingImage.Pix[idx+0] = newPixel.R
//			s.WorkingImage.Pix[idx+1] = newPixel.G
//			s.WorkingImage.Pix[idx+2] = newPixel.B
//			s.WorkingImage.Pix[idx+3] = newPixel.A
//			quant_error := subCol(inlineRGBACast(oldPixel.RGBA()), inlineRGBACast(newPixel.RGBA()))
//			// distribute the error to the neighbouring pixels
//			s.WorkingImage.At(x+1, y)
//
//		}
//
//	}
//}

func (s *State) ApplyFilters() {
	InfoLog("Applying filters")
	DebugLogf("Current filters: %+v", s.Filters) // %+v prints a struct with field names
	// set the shown image to the unmodified image
	s.WorkingImage.Pix = append([]uint8(nil), s.OrigImage.Pix...) // NOTE: found another copy by reference bug her
	if !bytes.Equal(s.WorkingImage.Pix, s.OrigImage.Pix) {
		FatalLog("Pixels copied incorrectly")
	}

	// for each filter, apply it to the shown image
	if s.Filters.IsGrayscaleEnabled {
		DebugLog("Grayscale filter applied")
		s.GrayscaleFilter()
	}
	if s.Filters.IsQuantizingEnabled {
		DebugLog("Quantizing applied")
		s.QuantizingFilter()
	}
	s.ShownImage = rl.NewImageFromImage(&s.WorkingImage)
}

func (s *State) Init() {
	InfoLog("Initialising state")
	s.Filters = Filters{
		IsGrayscaleEnabled:  false,
		IsDitheringEnabled:  false,
		DitheringLevel:      4,
		IsQuantizingEnabled: false,
		QuantizingBands:     50,
	}
	//load the image from the file
	s.ShownImage = rl.LoadImage("resources/image.png")

	//resize the image to fit the window on its largest axis
	aspectRatio := float32(s.ShownImage.Width) / float32(s.ShownImage.Height)

	// if it's longer on the x axis
	if aspectRatio > 1 {
		rl.ImageResizeNN(s.ShownImage, int32(rl.GetScreenWidth()), int32(float32(rl.GetScreenWidth())/aspectRatio))
	} else {
		// it's longer on the y axis
		rl.ImageResizeNN(s.ShownImage, int32(float32(rl.GetScreenHeight())*aspectRatio), int32(rl.GetScreenHeight()))
	}
	s.OrigImage = *s.ShownImage.ToImage().(*image.RGBA)
	s.WorkingImage = s.OrigImage

	//send the image to the GPU
	// rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
	s.BackgroundColour = rl.RayWhite
}
