package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const FilterCount = 5

const (
	English       Language = iota
	German                 = iota
	LanguageCount          = iota
)

type Theme int

const (
	ThemeLight Theme = iota
	ThemeDark
)

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
	SettingsWindow SettingsWindow

	RedHistogram   [256]int
	BlueHistogram  [256]int
	GreenHistogram [256]int

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters

	Config       Config
	LanguageData [LanguageCount]map[string]string
}

type Filters struct {
	IsGrayscaleEnabled bool

	IsQuantizingEnabled bool
	QuantizingBands     uint8

	IsDitheringEnabled           bool
	DitheringQuantizationBuckets uint8

	ChannelAdjustmentEnabled bool
	ChannelAdjustment        [3]float32

	IsBoxBlurEnabled  bool
	BoxBlurIterations int

	Order [FilterCount]string
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
	return strings.Join(MapOut(s.Filters.Order[:], Translate), ";")
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

func gaussianKernel(stDev float64, size int) [][]float64 {
	res := make([][]float64, size)
	for i := range size {
		res[i] = make([]float64, size)
	}
	sum := 0.0
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			exponent := -(float64(x*x+y*y) / (2 * stDev * stDev))
			coeff := 1 / (2 * math.Pi * stDev * stDev)
			res[x][y] = coeff * math.Exp(exponent)
			sum += res[x][y]
		}
	}
	//normalize
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			res[x][y] /= sum
		}
	}
	return res
}

func Clamp(v, min, max int) int {
	if v > max {
		return max
	}
	if v < min {
		return min
	}
	return v
}

func (s *State) BoxBlurFilter() {
	//kernelSize := 3
	w := s.WorkingImage.Bounds().Dx()
	h := s.WorkingImage.Bounds().Dy()
	for i := range s.Filters.BoxBlurIterations {
		_ = i
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				if x < 1 || y < 1 || x+1 == w || y+1 == h {
					continue
				}
				pixels := make([]color.Color, 9)
				pixels[0] = s.WorkingImage.At(x-1, y-1) // tl
				pixels[1] = s.WorkingImage.At(x, y-1)   // tm
				pixels[2] = s.WorkingImage.At(x+1, y-1) // tr

				pixels[3] = s.WorkingImage.At(x-1, y) // ml
				pixels[4] = s.WorkingImage.At(x, y)   // mm
				pixels[5] = s.WorkingImage.At(x+1, y) // mr

				pixels[6] = s.WorkingImage.At(x-1, y+1) // bl
				pixels[7] = s.WorkingImage.At(x, y+1)   // bm
				pixels[8] = s.WorkingImage.At(x+1, y+1) // br

				var rSum, gSum, bSum, aSum int
				for _, p := range pixels {
					r, g, b, a := p.RGBA()
					rSum += int(r >> 8)
					gSum += int(g >> 8)
					bSum += int(b >> 8)
					aSum += int(a >> 8)
				}
				rSum /= 9
				gSum /= 9
				bSum /= 9
				aSum /= 9
				s.WorkingImage.Set(x, y, color.RGBA{R: uint8(rSum), G: uint8(gSum), B: uint8(bSum), A: uint8(aSum)})
			}
		}
	}
}

func (s *State) GaussianBlurFilter() {
	//w := s.WorkingImage.Bounds().Dx()
	//h := s.WorkingImage.Bounds().Dy()
	//offset := s.Filters.GaussianKernelSize / 2
	//s.Filters.GaussianKernelSize = int(math.Pow(math.Ceil(6.0*s.Filters.GaussianDeviation), 2.0))
	//kernel := gaussianKernel(s.Filters.GaussianDeviation, s.Filters.GaussianKernelSize)
	//for y := 0; y < h; y++ {
	//	for x := 0; x < w; x++ {
	//		var rSum, gSum, bSum float64
	//
	//		// apply convolution
	//		for i := 0; i < s.Filters.GaussianKernelSize; i++ {
	//			for j := 0; j < s.Filters.GaussianKernelSize; j++ {
	//				px := Clamp(x+i-offset, 0, w)
	//				py := Clamp(y+i-offset, 0, h)
	//
	//				r, g, b, _ := s.WorkingImage.At(px, py).RGBA()
	//				rSum += float64(r) * kernel[i][j]
	//				gSum += float64(g) * kernel[i][j]
	//				bSum += float64(b) * kernel[i][j]
	//			}
	//		}
	//		blurredColor := color.RGBA{
	//			R: uint8(math.Min(math.Max(rSum/257, 0), 255)),
	//			G: uint8(math.Min(math.Max(gSum/257, 0), 255)),
	//			B: uint8(math.Min(math.Max(bSum/257, 0), 255)),
	//			A: 255,
	//		}
	//		s.WorkingImage.Set(x, y, blurredColor)
	//	}
	//}

}

//		func (s *State) GenerateNoiseImage(w, h int) {
//			DebugLog("Generating noise image")
//			seed := time.Now().Unix()
//			rng := rand.New(rand.NewSource(seed))
//			imageSlice := make([]uint8, w*h*4) // pixels * channels
//			for i := 0; i < w*h*4; i += 4 {
//				imageSlice[i+0] = uint8(rng.Intn(256))
//				imageSlice[i+1] = uint8(rng.Intn(256))
//				imageSlice[i+2] = uint8(rng.Intn(256))
//				imageSlice[i+3] = 255
//			}
//			rlImage := rl.NewImage(imageSlice, int32(w), int32(h), 0, rl.UncompressedR8g8b8a8).ToImage()
//			s.LoadImage(&rlImage)
//			s.RefreshImage()
//	}
func (s *State) RefreshImage() {
	// CONSTRUCT IMAGE PALETTE MAP
	s.ApplyFilters()                                             // up to 145ms
	s.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage) // >1ms
	// s.ConstructPalette()                                         // takes 120ms on full image but any quantization reduces a lot
	s.GenerateHistogram()
}

func (s *State) LoadImage(path string) {
	DebugLog("Loading image")
	// load the image from the file

	s.ShownImage = rl.LoadImage(path)
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

	// floor(x/bandWidth)*bandWidth + bandWidth/2
	// FIXME this is terrible, doesn't work and crashes in weird edge cases
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		s.WorkingImage.Pix[i+0] = QuantizeValue(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+0])
		s.WorkingImage.Pix[i+1] = QuantizeValue(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+1])
		s.WorkingImage.Pix[i+2] = QuantizeValue(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+2])
		s.WorkingImage.Pix[i+3] = s.OrigImage.Pix[i+3] // NOTE: forgot this line and image went invisible, write that down
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

func stdDev(vals []uint8) float64 {
	sum := 0.0
	for _, v := range vals {
		sum += float64(v)
	}
	mean := sum / float64(len(vals))
	dev := 0.0
	for _, v := range vals {
		dev += math.Pow(float64(v)-mean, 2)
	}

	dev /= float64(len(vals))
	return math.Sqrt(dev)
}

func RGBToHSV(r, g, b uint8) (float64, float64, float64) {
	rPrime := float64(r) / 255.0
	gPrime := float64(g) / 255.0
	bPrime := float64(b) / 255.0

	// value
	value := float64(max(rPrime, gPrime, bPrime))
	delta := value - float64(min(rPrime, gPrime, bPrime))

	// Hue

	// Saturation
	saturation := 0.0
	if value != 0 {
		saturation = float64(delta) / float64(value)
	}
	hue := 0.0
	switch value {
	case float64(rPrime):
		hue = 60.0 * float64(int((gPrime-bPrime)/delta)%6)
	case float64(gPrime):
		hue = 60.0 * (((bPrime - rPrime) / delta) + 2)
	case float64(bPrime):
		hue = 60.0 * (((rPrime - gPrime) / delta) + 4)
	}
	return hue, saturation, value
}

func hsvPixels(pix []uint8) []float64 {
	res := make([]float64, len(pix))
	for i := 0; i < len(pix); i += 4 {
		h, s, v := RGBToHSV(pix[i], pix[i+1], pix[i+2])
		res[i+0] = h
		res[i+1] = s
		res[i+2] = v
		res[i+3] = float64(pix[i+3])
	}
	return res
}

func everyNth[T any](vals []T, n, c int) []T {
	res := make([]T, len(vals)/n)
	for i := 0; i < len(vals)/n; i += 1 {
		res[i] = vals[n*i+c]
	}
	return res
}

// TODO: add a save & load menu with a drag and drop box

func (s *State) DitheringFilter() {
	bounds := s.WorkingImage.Bounds()
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			oldR, oldG, oldB, _ := s.WorkingImage.At(x, y).RGBA()
			newR := QuantizeValue(s.Filters.DitheringQuantizationBuckets, uint8(oldR))
			newG := QuantizeValue(s.Filters.DitheringQuantizationBuckets, uint8(oldG))
			newB := QuantizeValue(s.Filters.DitheringQuantizationBuckets, uint8(oldB))
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
		if s.Filters.IsGrayscaleEnabled && k == "control.grayscale" {
			t := time.Now()
			s.GrayscaleFilter()
			InfoLogf("Grayscale filter time: %v", time.Since(t))
		}
		if s.Filters.ChannelAdjustmentEnabled && k == "control.channeladjustment" {
			t := time.Now()
			s.TintFilter()
			InfoLogf("Tint filter time: %v", time.Since(t))
		}
		// SLOWish
		if s.Filters.IsQuantizingEnabled && k == "control.quantizing" {
			t := time.Now()
			s.QuantizingFilter()
			InfoLogf("Quant. filter time: %v", time.Since(t))
		}
		// SLOW
		if s.Filters.IsDitheringEnabled && k == "control.dithering" {
			t := time.Now()
			s.DitheringFilter()
			InfoLogf("Dither filter time: %v", time.Since(t))
		}
		if s.Filters.IsBoxBlurEnabled && k == "control.boxblur" {
			t := time.Now()
			s.BoxBlurFilter()
			InfoLogf("Gaussian filter time: %v", time.Since(t))
		}
	}
	s.ShownImage = rl.NewImageFromImage(&s.WorkingImage)
}

// TODO: logging not terminating colour escape codes
func (s *State) LoadLanguageData() {
	content, err := os.ReadFile("./resources/langs.json")
	if err != nil {
		FatalLogf("Couldn't read language data: %v", err.Error())
	}
	err = json.Unmarshal(content, &s.LanguageData)
	if err != nil {
		FatalLogf("Couldn't decode language file: %v", err.Error())
	}
}
func (s *State) Init() {
	InfoLog("Initialising state")
	s.Filters = Filters{
		DitheringQuantizationBuckets: 190,
		QuantizingBands:              50,
		ChannelAdjustment:            [3]float32{1.0, 1.0, 1.0},
		BoxBlurIterations:            3,
		Order:                        [FilterCount]string{"control.grayscale", "control.quantizing", "control.dithering", "control.channeladjustment", "control.boxblur"}, // initial Order
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
	s.SettingsWindow = SettingsWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}
	s.LoadLanguageData()
	s.QuantizationKMeansIterations = 10 // adjust for perfk
	s.LoadImage("./resources/image.png")
	//load the image from the file

	//resize the image to fit the window on its largest axis
	//send the image to the GPU
	// rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
	s.BackgroundColour = rl.RayWhite
}

func (s *State) SaveImage() {
	extension := string(s.Config.FileFormat)
	InfoLogf("Saving as %s format", extension)
	f, err := os.Create("./resources/output." + extension)
	if err != nil {
		FatalLogf("Couldn't create file: %v", err.Error())
	}
	switch s.Config.FileFormat {
	case PNG:
		err = png.Encode(f, image.Image(&s.WorkingImage))
	case TIFF:
		err = png.Encode(f, image.Image(&s.WorkingImage))
	case BMP:
		err = png.Encode(f, image.Image(&s.WorkingImage))
	case JPG:
		err = png.Encode(f, image.Image(&s.WorkingImage))
	}
	if err != nil {
		FatalLogf("Couldn't encode image: %v", err.Error())
	}
}

func (s *State) SaveState() {
	InfoLog("Saving state")
	f, _ := os.Create("./resources/state.json")
	c, _ := json.Marshal(state) // what a stupid API
	_, _ = f.Write(c)
}

func (s *State) Close() {
	state.SaveImage()
	state.SaveState()
	rl.CloseWindow()
	os.Exit(0)
}
