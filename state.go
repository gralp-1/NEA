package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"slices"
	"strings"
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

const FilterCount = 6

const (
	English       Language = iota
	German                 = iota
	LanguageCount          = iota
)

type State struct {
	ImageLoaded bool
	ImagePath   string

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage image.RGBA
	ShownImage   *rl.Image
	ImagePalette []rl.Color

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

	LightenDarken float64

	Order [FilterCount]string
}

type ColourHistogram struct {
	RedChannel   map[uint8]int
	GreenChannel map[uint8]int
	BlueChannel  map[uint8]int
}

func (s *State) LightenDarken() {
	// +1 makes everything white
	// -1 makes everything black
	// 0 does nothing
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		s.WorkingImage.Pix[i+0] = uint8(Clamp(float64(s.WorkingImage.Pix[i+0])*(s.Filters.LightenDarken+1), 0, 255))
		s.WorkingImage.Pix[i+1] = uint8(Clamp(float64(s.WorkingImage.Pix[i+1])*(s.Filters.LightenDarken+1), 0, 255))
		s.WorkingImage.Pix[i+2] = uint8(Clamp(float64(s.WorkingImage.Pix[i+2])*(s.Filters.LightenDarken+1), 0, 255))
	}
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

func Clamp[T interface{ ~int | ~float64 }](v, min, max T) T {
	if v >= max {
		return max
	}
	if v <= min {
		return min
	}
	return v
}

func (s *State) BoxBlurFilter() {
	//kernelSize := 3
	bounds := s.WorkingImage.Bounds()
	// temp := image.NewRGBA(bounds)
	w, h := bounds.Dx(), bounds.Dy()
	// rad := s.Filters.BoxBlurIterations
	// // horizontal pass
	// for y := 0; y < h; y++ {
	// 	for x := 0; x < w; x++ {
	// 		count := 0
	// 		var r, g, b uint32
	// 		for i := -rad; i <= rad; i++ {
	// 			px := Clamp(x+int(i), 0, w-1)
	// 			r, g, b, _ = s.OrigImage.At(px, y).RGBA()
	// 			xi := x + int(i)
	// 			if xi >= 0 && xi < w {
	// 				c := s.OrigImage.At(xi, y)
	// 				cr, cg, cb, _ := c.RGBA()
	// 				r += (cr >> 8)
	// 				g += (cg >> 8)
	// 				b += (cb >> 8)
	// 				count++
	// 			}
	// 		}
	// 		temp.SetRGBA(x, y, color.RGBA{
	// 			R: uint8(int(r) / count),
	// 			G: uint8(int(g) / count),
	// 			B: uint8(int(b) / count),
	// 			A: 255,
	// 		})
	// 	}
	// }
	// // vertical pass
	// for y := 0; y < h; y++ {
	// 	for x := 0; x < w; x++ {
	// 		count := 0
	// 		var r, g, b uint32
	// 		for i := -rad; i <= rad; i++ {
	// 			py := Clamp(y+int(i), 0, h-1)
	// 			r, g, b, _ = s.OrigImage.At(x, py).RGBA()
	// 			yi := y + int(i)
	// 			if yi >= 0 && yi < h {
	// 				c := s.OrigImage.At(x, yi)
	// 				cr, cg, cb, _ := c.RGBA()
	// 				r += (cr >> 8)
	// 				g += (cg >> 8)
	// 				b += (cb >> 8)
	// 				count++
	// 			}
	// 		}
	// 		temp.SetRGBA(x, y, color.RGBA{
	// 			R: uint8(int(r) / count),
	// 			G: uint8(int(g) / count),
	// 			B: uint8(int(b) / count),
	// 			A: 255,
	// 		})
	// 	}
	// }
	// s.WorkingImage = *temp
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
	// s.ConstructPalette()                                      // takes 120ms on full image but any quantization reduces a lot
	s.GenerateHistogram()
}

func (s *State) LoadImageFile(path string) {
	// get the extension and match it, if there's an error in this we just return without settings ImageLoaded to true
	fileParts := strings.Split(path, ".")
	extension := strings.ToLower(fileParts[len(fileParts)-1])
	f, err := os.Open(s.ImagePath)
	if err != nil {
		return
	}

	var image image.Image

	if extension == "jpg" || extension == "jpeg" {
		image, err = jpeg.Decode(f)
	}
	if extension == "png" {
		image, err = png.Decode(f)
	}
	if extension == "tiff" {
		image, err = tiff.Decode(f)
	}
	if extension == "bmp" {
		image, err = bmp.Decode(f)
	}

	if err != nil {
		return
	}
	s.LoadImage(image)
	s.ImageLoaded = true
}

func (s *State) LoadImage(img image.Image) {
	DebugLog("Loading image")
	// load the image from the file

	s.ShownImage = rl.NewImageFromImage(img)
	// aspectRatio := float32(s.ShownImage.Width) / float32(s.ShownImage.Height)

	// // if it's longer on the x axis
	// if aspectRatio > 1 {
	// 	rl.ImageResizeNN(s.ShownImage, int32(rl.GetScreenWidth()), int32(float32(rl.GetScreenWidth())/aspectRatio))
	// } else {
	// 	// it's longer on the y axis
	// 	rl.ImageResizeNN(s.ShownImage, int32(float32(rl.GetScreenHeight())*aspectRatio), int32(rl.GetScreenHeight()))
	// }
	s.OrigImage = *s.ShownImage.ToImage().(*image.RGBA)
	s.WorkingImage = s.OrigImage
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))
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
		s.WorkingImage.Pix[i+0] = Quantize(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+0])
		s.WorkingImage.Pix[i+1] = Quantize(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+1])
		s.WorkingImage.Pix[i+2] = Quantize(state.Filters.QuantizingBands, s.WorkingImage.Pix[i+2])
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
			newR := Quantize(s.Filters.DitheringQuantizationBuckets, uint8(oldR))
			newG := Quantize(s.Filters.DitheringQuantizationBuckets, uint8(oldG))
			newB := Quantize(s.Filters.DitheringQuantizationBuckets, uint8(oldB))
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
		if k == "control.lightendarken" {
			t := time.Now()
			s.LightenDarken()
			InfoLogf("Lighten/darken filter time: %v", time.Since(t))
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
	// Image loading will be called from main when file is drag&dropped
	InfoLog("Initialising state")
	s.LoadFonts()
	InfoLog("Initialising filters")
	s.Filters = Filters{
		DitheringQuantizationBuckets: 190,
		QuantizingBands:              50,
		ChannelAdjustment:            [3]float32{1.0, 1.0, 1.0},
		BoxBlurIterations:            3,
		LightenDarken:                0.0,
		Order:                        [FilterCount]string{"control.grayscale", "control.quantizing", "control.dithering", "control.channeladjustment", "control.boxblur", "control.lightendarken"}, // initial Order
	}
	InfoLog("Initialising windows")
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
		Showing:                  false,
		Anchor:                   rl.Vector2{X: 20, Y: 20},
		IsFileTypeDropDownActive: false,
	}
	s.SettingsWindow = SettingsWindow{
		Showing: false,
		Anchor:  rl.Vector2{X: 20, Y: 20},
	}

	InfoLog("Initialising language data")
	s.LoadLanguageData()

	s.Config.FileFormat = TIFF
	InfoLog("Initialising font size")
	s.Config.FontSize = 10
	s.SetFontSize()
	InfoLog("Finished state init")

	//resize the image to fit the window on its largest axis
	//send the image to the GPU
	// rl.UpdateTexture(s.CurrentTexture, Uint8SliceToRGBASlice(s.OrigImage.Pix))
	// state.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)

	//initialise everything else
}

func (s *State) SaveImage() {
	extension := s.Config.GetActiveFileFormat().String()
	InfoLogf("Saving as output.%s", extension)
	f, err := os.Create("./output." + extension)
	if err != nil {
		FatalLogf("Couldn't create file: %v", err.Error())
	}
	switch s.Config.FileFormat {
	case PNG:
		err = png.Encode(f, image.Image(&s.WorkingImage))
	case TIFF:
		err = tiff.Encode(f, image.Image(&s.WorkingImage), &tiff.Options{Compression: tiff.Uncompressed, Predictor: true})
	case BMP:
		err = bmp.Encode(f, image.Image(&s.WorkingImage))
	case JPG:
		err = jpeg.Encode(f, image.Image(&s.WorkingImage), &jpeg.Options{Quality: 100})
	}
	if err != nil {
		FatalLogf("Couldn't encode image: %v", err.Error())
	}
}

func (s *State) Close() {
	state.SaveImage()
	rl.CloseWindow()
	os.Exit(0)
}

var DarkThemeData [][]int = [][]int{
	{0, 0, 0x878787ff},  // DEFAULT_BORDER_COLOR_NORMAL
	{0, 1, 0x2c2c2cff},  // DEFAULT_BASE_COLOR_NORMAL
	{0, 2, 0xc3c3c3ff},  // DEFAULT_TEXT_COLOR_NORMAL
	{0, 3, 0xe1e1e1ff},  // DEFAULT_BORDER_COLOR_FOCUSED
	{0, 4, 0x848484ff},  // DEFAULT_BASE_COLOR_FOCUSED
	{0, 5, 0x181818ff},  // DEFAULT_TEXT_COLOR_FOCUSED
	{0, 6, 0x000000ff},  // DEFAULT_BORDER_COLOR_PRESSED
	{0, 7, 0xefefefff},  // DEFAULT_BASE_COLOR_PRESSED
	{0, 8, 0x202020ff},  // DEFAULT_TEXT_COLOR_PRESSED
	{0, 9, 0x6a6a6aff},  // DEFAULT_BORDER_COLOR_DISABLED
	{0, 10, 0x818181ff}, // DEFAULT_BASE_COLOR_DISABLED
	{0, 11, 0x606060ff}, // DEFAULT_TEXT_COLOR_DISABLED
	{0, 18, 0x9d9d9dff}, // DEFAULT_LINE_COLOR
	{0, 19, 0x3c3c3cff}, // DEFAULT_BACKGROUND_COLOR
	{1, 5, 0xf7f7f7ff},  // LABEL_TEXT_COLOR_FOCUSED
	{1, 8, 0x898989ff},  // LABEL_TEXT_COLOR_PRESSED
	{4, 5, 0xb0b0b0ff},  // SLIDER_TEXT_COLOR_FOCUSED
	{5, 5, 0x848484ff},  // PROGRESSBAR_TEXT_COLOR_FOCUSED
	{9, 5, 0xf5f5f5ff},  // TEXTBOX_TEXT_COLOR_FOCUSED
	{10, 5, 0xf6f6f6ff}, // VALUEBOX_TEXT_COLOR_FOCUSED
}

func (s *State) ChangeTheme() {
	switch s.Config.CurrentTheme {
	case ThemeLight:
		gui.LoadStyleDefault()
	case ThemeDark:
		for _, control := range DarkThemeData {
			gui.SetStyle(int32(control[0]), int32(control[1]), int64(control[2]))
		}
	}
	s.LoadFonts()
	s.ChangeFont()
}

var FontLibrary [int(FontCount)]rl.Font

func (s *State) LoadFonts() {
	DebugLogf("Loading %d fonts", FontCount)
	FontLibrary[FontDefault] = rl.GetFontDefault()
	FontLibrary[FontArial] = rl.LoadFont("resources/arial.ttf")
	FontLibrary[FontBerkleyMono] = rl.LoadFont("resources/berkley_mono.otf")
	FontLibrary[FontComicSans] = rl.LoadFont("resources/comic_sans.ttf")
	FontLibrary[FontZapfino] = rl.LoadFont("resources/zapfino.ttf")
	FontLibrary[FontSpleen] = rl.LoadFont("resources/spleen.otf")
}
func (s *State) ChangeFont() {
	DebugLogf("Changing font to %d", s.Config.CurrentFont)
	gui.SetFont(FontLibrary[s.Config.CurrentFont])
}
func (s *State) SetFontSize() {
	gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, s.Config.FontSize)
}
