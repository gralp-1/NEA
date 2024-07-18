package main

import (
	"image"
	"image/color"
	"log"
	"time"
	"unsafe"

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

// TODO: filter diff so we can incrementally apply filters instead of every frame

func (s *State) Init() {
	log.Print("Initialising state")
	// load the image from the file
	s.ShownImage = rl.LoadImage("resources/image.png")

	// resize the image to fit the window on its largest axis
	aspectRatio := float32(s.ShownImage.Width) / float32(s.ShownImage.Height)

	// if it's longer on the x axis
	if aspectRatio > 1 {
		rl.ImageResizeNN(s.ShownImage, int32(rl.GetScreenWidth()), int32(float32(rl.GetScreenWidth())/aspectRatio))
	} else {
		// it's longer on the y axis
		rl.ImageResizeNN(s.ShownImage, int32(float32(rl.GetScreenHeight())*aspectRatio), int32(rl.GetScreenHeight()))
	}
	s.OrigImage = s.ShownImage.ToImage().(*image.RGBA)

	// send the image to the GPU
	length := s.ShownImage.Width * s.ShownImage.Height
	slice := (*[1 << 30]color.RGBA)(unsafe.Pointer(state.ShownImage))[:length:length]
	rl.UpdateTexture(s.CurrentTexture, slice)
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	// initialise everything else
	s.BackgroundColour = rl.RayWhite
}

func TimeFunc(x func(any) any, args any) (any, time.Duration) {
	start := time.Now()

	return x(args), time.Since(start)
}

func (s *State) ApplyFilters() {
	log.Print("Applying filters")
	// set the shown image to the unmodified image
	// time this line
	start := time.Now()

	s.ShownImage = rl.NewImageFromImage(s.OrigImage)
	duration := time.Since(start)
	log.Printf("Time taken to copy image: %v", duration.Milliseconds())

	tempImage := s.OrigImage

	// for each filter, apply it to the shown image
	if s.Filters.IsGrayscaleEnabled {
		_, duration := TimeFunc(rl.ImageColorGrayscale, s.ShownImage)
		log.Printf("Time taken to apply grayscale: %v", duration.Milliseconds())
		// rl.ImageColorGrayscale(s.ShownImage)
	}
}
