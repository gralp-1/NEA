package main

import (
	"image"
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage  *image.RGBA
	ShownImage *rl.Image

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
}

type Filters struct {
	IsGrayscaleEnabled bool
	IsDitheringEnabled bool
	DitheringLevel     int
}

func Uint8SliceToRGBASlice(slice []uint8) []color.RGBA {
	res := make([]color.RGBA, len(slice)/4)
	for i := 0; i < len(slice); i += 4 {
		res[i/4] = color.RGBA{slice[i], slice[i+1], slice[i+2], slice[i+3]}
	}
	return res
}

func (s *State) GrayscaleFilter() {
	res := make([]uint8, len(s.OrigImage.Pix))
	for i := 0; i < len(s.OrigImage.Pix); i += 4 {
		mean := uint8((int(s.OrigImage.Pix[i]) + int(s.OrigImage.Pix[i+1]) + int(s.OrigImage.Pix[i+2])) / 3)
		res[i+0] = mean
		res[i+1] = mean
		res[i+2] = mean
		res[i+3] = s.OrigImage.Pix[i+3] // NOTE: forgot this line and image went invisible, write that down
	}
	tempImage := s.OrigImage
	tempImage.Pix = res
	s.ShownImage = rl.NewImageFromImage(tempImage)
}

// TODO: filter diff so we can incrementally apply filters instead of every frame

// NOTE: can't be used on the critical path, too slow
func (s *State) ApplyFilters() {
	InfoLog("Applying filters")
	DebugLogf("Current filters: %+v", s.Filters) // %+v prints a struct with field names
	// set the shown image to the unmodified image

	s.ShownImage = rl.NewImageFromImage(s.OrigImage) // ~100ms
	if *s.ShownImage != *rl.NewImageFromImage(s.OrigImage) {
		ErrorLog("Image reset failed")
		ErrorLogf("RHS: %+v", s.ShownImage.Data)
		FatalLogf("RHS: %+v", rl.NewImageFromImage(s.OrigImage).Data)
	}

	// for each filter, apply it to the shown image
	if s.Filters.IsGrayscaleEnabled {
		DebugLog("Grayscale filter applied")
		s.GrayscaleFilter()
	}
}

func (s *State) Init() {
	InfoLog("Initialising state")
	s.Filters = Filters{
		IsGrayscaleEnabled: false,
		IsDitheringEnabled: false,
		DitheringLevel:     0,
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
	s.OrigImage = s.ShownImage.ToImage().(*image.RGBA)

	//send the image to the GPU
	rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
	s.BackgroundColour = rl.RayWhite
}
