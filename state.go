package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const FilterCount = 4

type State struct {

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage                    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage                 image.RGBA
	ShownImage                   *rl.Image
	ImagePalette                 []rl.Color
	QuantizationKMeansIterations int

	FilterWindow   FilterOrderWindow
	PaletteWindow  PaletteWindow
	HelpWindow     HelpWindow
	SaveLoadWindow SaveLoadWindow

	RedHistogram   [256]int
	BlueHistogram  [256]int
	GreenHistogram [256]int

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
}

type Filters struct {
	IsGrayscaleEnabled           bool
	IsQuantizingEnabled          bool
	QuantizingBands              uint8
	IsDitheringEnabled           bool
	DitheringQuantizationBuckets uint8
	ChannelAdjustmentEnabled     bool
	ChannelAdjustment            [3]float32
	Order                        [FilterCount]string
}

type ColourHistogram struct {
	RedChannel   map[uint8]int
	GreenChannel map[uint8]int
	BlueChannel  map[uint8]int
}

func (h *ColourHistogram) ConstructHistogram() {
	h.RedChannel = make(map[uint8]int, 256)
	h.GreenChannel = make(map[uint8]int, 256)
	h.BlueChannel = make(map[uint8]int, 256)
	for y := 0; y < state.WorkingImage.Rect.Dy(); y++ {
		for x := 0; x < state.WorkingImage.Rect.Dx(); x++ {
			r, g, b, _ := state.WorkingImage.At(x, y).RGBA()
			h.RedChannel[uint8(r)] += 1
			h.GreenChannel[uint8(g)] += 1
			h.BlueChannel[uint8(b)] += 1
		}
	}
}

func pixelToRlColor(col color.RGBA) rl.Color { // inline cast
	r, g, b, a := col.RGBA()
	return rl.Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}
}

func (s *State) ConstructPalette() {
	pal := make(map[rl.Color]float64)
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		r := s.WorkingImage.Pix[i]
		g := s.WorkingImage.Pix[i+1]
		b := s.WorkingImage.Pix[i+2]
		pal[rl.Color{R: r, G: g, B: b, A: 255}] = 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
	}
	colours := make([]rl.Color, 0, len(pal))
	for k, _ := range pal {
		colours = append(colours, k)
	}
	slices.SortFunc(colours, func(a, b rl.Color) int {
		if pal[a] > pal[b] {
			return -1
		}
		return 1
	})
	state.ImagePalette = colours
}

func (s *State) GetFiltersListViewString() string {
	return strings.Join(s.Filters.Order[:], ";")
}

func (s *State) GetImageColours() []rl.Color {
	pixels := make([]rl.Color, len(s.OrigImage.Pix)/4)
	for idx := 0; idx < len(s.OrigImage.Pix)/4; idx += 4 {
		pixels[idx] = rl.Color{R: s.OrigImage.Pix[idx], G: s.OrigImage.Pix[idx+1], B: s.OrigImage.Pix[idx+2], A: s.OrigImage.Pix[idx+3]}
	}
	return pixels
}
func (s *State) GenerateHistogram() {
	s.RedHistogram = [256]int{}
	s.GreenHistogram = [256]int{}
	s.BlueHistogram = [256]int{}
	for _, col := range PixSliceToColourSlice(state.WorkingImage.Pix) {
		s.RedHistogram[col.R]++
		s.GreenHistogram[col.G]++
		s.BlueHistogram[col.B]++
	}
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
	rlImage := rl.NewImage(imageSlice, int32(w), int32(h), 0, rl.UncompressedR8g8b8a8).ToImage()
	s.LoadImage(&rlImage)
	s.RefreshImage()
}
func (s *State) RefreshImage() {
	// CONSTRUCT IMAGE PALETTE MAP
	start := time.Now()
	s.ApplyFilters()
	s.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage) // ~100ms
	s.ConstructPalette()
	s.GenerateHistogram()
	elapsed := time.Since(start)
	DebugLogf("Refreshing image took %v", elapsed.String())
}

func (s *State) LoadImage(img *image.Image) {
	DebugLog("Loading image")
	// load the image from the file
	s.ShownImage = rl.NewImageFromImage(*img)
	s.OrigImage = *(*img).(*image.RGBA) // WHAT THE FUCK
	s.WorkingImage = s.OrigImage

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

// TODO: add a save & load menu with a drag and drop box

func (s *State) quantizeValue(v uint8) uint8 {
	// TODO: make some adjustable curve??
	bandCount := s.Filters.DitheringQuantizationBuckets
	bandWidth := uint8(math.Floor(255 / float64(bandCount)))
	for i := range bandCount + 1 {
		if (bandCount-i)*bandWidth < v {
			return bandWidth * (bandCount - i)
		}
	}
	return 0
}
func (s *State) DitheringFilter() {
	bounds := s.WorkingImage.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			oldR, oldG, oldB, _ := s.WorkingImage.At(x, y).RGBA()
			newR := s.quantizeValue(uint8(oldR)) // TODO: This is too aggressive
			newG := s.quantizeValue(uint8(oldG))
			newB := s.quantizeValue(uint8(oldB))
			s.WorkingImage.SetRGBA(x, y, color.RGBA{R: newR, G: newG, B: newB, A: 255})

			errR := uint8(oldR) - newR
			errG := uint8(oldG) - newG
			errB := uint8(oldB) - newB
			distributeError(&s.WorkingImage, x+1, y+0, errR, errG, errB, 7.0/16.0) // direct right
			distributeError(&s.WorkingImage, x-1, y+1, errR, errG, errB, 3.0/16.0) // top left
			distributeError(&s.WorkingImage, x+0, y+1, errR, errG, errB, 5.0/16.0) // top
			distributeError(&s.WorkingImage, x+1, y+1, errR, errG, errB, 1.0/16.0) // bottom right
		}
	}
}

func (s *State) ApplyFilters() {
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
			s.DitheringFilter()
		}
	}
	s.ShownImage = rl.NewImageFromImage(&s.WorkingImage)
}

func (s *State) Init() {
	InfoLog("Initialising state")
	s.Filters = Filters{
		IsGrayscaleEnabled:           false,
		IsDitheringEnabled:           false,
		DitheringQuantizationBuckets: 190,
		IsQuantizingEnabled:          false,
		QuantizingBands:              50,
		ChannelAdjustmentEnabled:     false,
		ChannelAdjustment:            [3]float32{1.0, 1.0, 1.0},
		Order:                        [4]string{"Grayscale", "Quantizing", "Dithering", "Tint"}, // initial Order
	}
	s.FilterWindow = FilterOrderWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.PaletteWindow = PaletteWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.HelpWindow = HelpWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.SaveLoadWindow = SaveLoadWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.QuantizationKMeansIterations = 10 // adjust for perf
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
	s.ConstructPalette()

	//send the image to the GPU
	// rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
	s.BackgroundColour = rl.RayWhite
}

func (s *State) SaveImage() {
	f, _ := os.Create("image.png")
	_ = png.Encode(f, state.WorkingImage.SubImage(state.WorkingImage.Rect)) // what a stupid API
}
