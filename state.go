package main

import (
	"bytes"
	"image"
	"math"
	"math/rand"
	"strings"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const FILTER_COUNT = 4

type State struct {

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage                    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage                 image.RGBA
	ShownImage                   *rl.Image
	ImagePalette                 []rl.Color
	QuantizationKmeansIterations int
	FilterWindow                 FilterOrderWindow

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
}

type FilterOrderWindow struct {
	Showing     bool
	Anchor      rl.Vector2
	ScrollIndex int32
	Active      int32
}

func (f *FilterOrderWindow) New() FilterOrderWindow {
	return FilterOrderWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 150, Y: 150},
	}
}
func (f *FilterOrderWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(f.Anchor.X, f.Anchor.Y, 300, 260)
}
func (f *FilterOrderWindow) Draw() {
	f.Showing = !gui.WindowBox(f.getRect(), "Filter Order Configuration")
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+30, 100, 10), "Applied first")
	f.Active = gui.ListView(
		rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+50, 100, 180),
		state.Filters.GetFiltersListViewString(),
		&f.ScrollIndex,
		f.Active,
	)
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+240, 100, 10), "Applied last")
	// Promote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+30, 100, 50), "Promote Selected") {
		f.Promote()
	}
	// Demote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+80, 100, 50), "Demote Selected") {
		f.Demote()
	}
}
func (f *FilterOrderWindow) Promote() {
	if f.Active == 0 {
		DebugLog("Attempted to promote first index")
	} else {
		state.Filters.Order[f.Active], state.Filters.Order[f.Active-1] = state.Filters.Order[f.Active-1], state.Filters.Order[f.Active]
		f.Active--
		state.RefreshImage()
	}
}
func (f *FilterOrderWindow) Demote() {
	if f.Active == FILTER_COUNT-1 {
		DebugLog("Attempted to demote last index")
	} else {
		state.Filters.Order[f.Active], state.Filters.Order[f.Active+1] = state.Filters.Order[f.Active+1], state.Filters.Order[f.Active]
		f.Active++
		state.RefreshImage()
	}
}

type Filters struct {
	IsGrayscaleEnabled       bool
	IsQuantizingEnabled      bool
	QuantizingBands          uint8
	IsDitheringEnabled       bool
	DitheringLevel           int
	ChannelAdjustmentEnabled bool
	ChannelAdjustment        [3]float32
	Order                    [FILTER_COUNT]string
}

func (f *Filters) GetFiltersListViewString() string {
	return strings.Join(f.Order[:], ";")
}

func (s *State) GetImageColours() []rl.Color {
	pixels := make([]rl.Color, len(s.OrigImage.Pix)/4)
	for idx := 0; idx < len(s.OrigImage.Pix)/4; idx += 4 {
		pixels[idx] = rl.Color{R: s.OrigImage.Pix[idx], G: s.OrigImage.Pix[idx+1], B: s.OrigImage.Pix[idx+2], A: s.OrigImage.Pix[idx+3]}
	}
	return pixels
}

// TODO: move to utils
func removeDuplicates[T comparable](s []T) []T {
	seen := make(map[T]bool)
	var result []T
	for _, v := range s {
		if !seen[v] {
			result = append(result, v)
			seen[v] = true
		}
	}
	return result
}

func (s *State) GenerateNoiseImage(w, h int) {
	DebugLog("Generating noise image")
	seed := time.Now().Unix()
	rng := rand.New(rand.NewSource(seed))
	imageSlice := make([]uint8, w*h*4) // pixels * channels
	for i := 0; i < w*h*4; i += 4 {
		imageSlice[i+0] = uint8(rng.Intn(256))
		imageSlice[i+1] = uint8(rng.Intn(256))
		imageSlice[i+2] = uint8(rng.Intn(256))
		imageSlice[i+3] = 255
	}
	rlImage := rl.NewImage(imageSlice, int32(w), int32(h), 0, rl.UncompressedR8g8b8a8)
	s.LoadImage(rlImage)
	s.RefreshImage()
}
func (s *State) RefreshImage() {
	start := time.Now()
	s.ApplyFilters()
	s.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage) // ~100ms
	elapsed := time.Since(start)
	DebugLogf("Refreshing image took %v", elapsed.String())
}

func (s *State) LoadImage(img *rl.Image) {
	DebugLog("Loading image")
	// load the image from the file
	s.ShownImage = img

	// resize the image to fit the window on its largest axis
	// aspectRatio := float64(s.ShownImage.Width) / float64(s.ShownImage.Height)
	// TODO: figure out how to resize
	// if it's longer on the x axis
	// if aspectRatio > 1.0 {
	// 	rl.ImageResizeNN(s.ShownImage, int32(rl.GetScreenWidth()), int32(float64(rl.GetScreenWidth())/aspectRatio))
	// } else {
	// 	// it's longer on the y axis
	// 	rl.ImageResizeNN(s.ShownImage, int32(float64(rl.GetScreenHeight())*aspectRatio), int32(rl.GetScreenHeight()))
	// }
	s.OrigImage = *s.ShownImage.ToImage().(*image.RGBA)
	s.WorkingImage = s.OrigImage

}
func (s *State) GenerateImagePalette() {
	DebugLog("Generating Image palette")
	// for each image
	s.ImagePalette = removeDuplicates(s.GetImageColours())
	DebugLogf("Image Pallete Length: %v", len(s.ImagePalette))
	DebugLogf("Unique colour ratio: %f%%", (float64(len(s.ImagePalette))/float64(len(s.OrigImage.Pix)/4))*100)
}

func (s *State) GrayscaleFilter() {
	DebugLog("Grayscale filter applied")
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		mean := uint8((int(s.WorkingImage.Pix[i]) + int(s.WorkingImage.Pix[i+1]) + int(s.WorkingImage.Pix[i+2])) / 3)
		s.WorkingImage.Pix[i+0] = mean
		s.WorkingImage.Pix[i+1] = mean
		s.WorkingImage.Pix[i+2] = mean
	}
}

func (s *State) QuantizingFilter() {
	DebugLog("Quantizing filter applied")
	// maybe preserve current code and make a "simple" and "accurate" mode
	// this is an NP-hard algorithm as we need to ca
	// We need to do k means clustering on the pixel space
	// then set each colour in the local space to the mean
	// value of the space... I think

	quantizationBandWidth := 255 / float64(state.Filters.QuantizingBands-1)
	// floor(x/bandWidth)*bandWidth + bandWidth/2
	// FIXME this is terrible, doesn't work and crashes in weird edge cases
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

func (s *State) TintFilter() {
	DebugLog("Tint filter applied")
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		s.WorkingImage.Pix[i+0] = uint8(float32(s.WorkingImage.Pix[i+0]) * s.Filters.ChannelAdjustment[0])
		s.WorkingImage.Pix[i+1] = uint8(float32(s.WorkingImage.Pix[i+1]) * s.Filters.ChannelAdjustment[1])
		s.WorkingImage.Pix[i+2] = uint8(float32(s.WorkingImage.Pix[i+2]) * s.Filters.ChannelAdjustment[2])
	}
}
func (s *State) ApplyFilters() {
	// TODO: make sure to do them in Order
	InfoLog("Applying filters")
	DebugLogf("Current filters: %+v", s.Filters) // %+v prints a struct with field names
	// set the shown image to the unmodified image
	s.WorkingImage.Pix = append([]uint8(nil), s.OrigImage.Pix...) // NOTE: found another copy by reference bug her
	if !bytes.Equal(s.WorkingImage.Pix, s.OrigImage.Pix) {
		FatalLog("Pixels copied incorrectly")
	}

	for _, k := range s.Filters.Order {

		// for each filter, apply it to the shown image
		if s.Filters.IsGrayscaleEnabled && k == "Grayscale" {
			s.GrayscaleFilter()
		}
		if s.Filters.ChannelAdjustmentEnabled && k == "Tint" {
			s.TintFilter()
		}
		if s.Filters.IsQuantizingEnabled && k == "Quantizing" {
			s.QuantizingFilter()
		}
		if s.Filters.IsDitheringEnabled && k == "Dithering" {
			s.Dither()
		}
	}
	s.ShownImage = rl.NewImageFromImage(&s.WorkingImage)
}
func (s *State) Dither() {
	InfoLog("Attempted to dither filter")
	rl.ImageDither(s.ShownImage, 5, 5, 5, 5)
}

// TODO: add labels to the top and bottom to show first and last initialized
// TODO: make sure the features in the filter change window line up properly
func (s *State) Init() {
	InfoLog("Initialising state")
	s.Filters = Filters{
		IsGrayscaleEnabled:       false,
		IsDitheringEnabled:       false,
		DitheringLevel:           4,
		IsQuantizingEnabled:      false,
		QuantizingBands:          50,
		ChannelAdjustmentEnabled: false,
		ChannelAdjustment:        [3]float32{1.0, 1.0, 1.0},
		Order:                    [4]string{"Grayscale", "Quantizing", "Dithering", "Tint"}, // initial Order
	}
	s.FilterWindow = FilterOrderWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.QuantizationKmeansIterations = 10 // adjust for perf
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

	// generate the image colour palette
	s.GenerateImagePalette()

	//send the image to the GPU
	// rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
	s.BackgroundColour = rl.RayWhite
}
