package main

import (
	"bytes"
	"image"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	set "github.com/golang-collections/collections/set"
)

type State struct {

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage                    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage                 image.RGBA
	ShownImage                   *rl.Image
	ImagePalette                 set.Set
	QuantizationKmeansIterations int

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
}

type Filters struct {
	IsGrayscaleEnabled       bool
	IsQuantizingEnabled      bool
	QuantizingBands          uint8
	IsDitheringEnabled       bool
	DitheringLevel           int
	ChannelAdjustmentEnabled bool
	ChannelAdjustment        [3]float32
}

func (s *State) GetImageColours() []rl.Color {
	pixels := make([]rl.Color, len(s.OrigImage.Pix)/4, len(s.OrigImage.Pix)/4)
	for idx := range s.OrigImage.Pix {
		pixels[idx/4] = rl.Color{R: s.OrigImage.Pix[idx], G: s.OrigImage.Pix[idx+1], B: s.OrigImage.Pix[idx+2], A: s.OrigImage.Pix[idx+3]}
		idx += 4
	}
	return pixels
}
func (s *State) GenerateImagePalette() {
	// for each image
	s.ImagePalette = *set.New(s.GetImageColours())
}

func (s *State) GrayscaleFilter() {
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		mean := uint8((int(s.WorkingImage.Pix[i]) + int(s.WorkingImage.Pix[i+1]) + int(s.WorkingImage.Pix[i+2])) / 3)
		s.WorkingImage.Pix[i+0] = mean
		s.WorkingImage.Pix[i+1] = mean
		s.WorkingImage.Pix[i+2] = mean
	}
}

type Integer interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | // unsigned
		~int | ~int8 | ~int16 | ~int32 | ~int64 // signed
}

type ClusteredColour struct {
	colour  rl.Color
	cluster int
}

func Dist[T Integer](args ...T) float64 {
	var sum float64 = 0
	for _, v := range args {
		sum += float64(v * v)
	}
	return math.Sqrt(sum)
}

func Nearest(p ClusteredColour, mean []rl.Color) (int, float64) {
	iMin := 0
	dMin := Dist(p.colour.R-mean[0].R, p.colour.G-mean[0].G, p.colour.B-mean[0].B)
	for i := 1; i < len(mean); i++ {
		d := Dist(p.colour.R-mean[i].R, p.colour.G-mean[i].G, p.colour.B-mean[i].B)
		if d < dMin {
			dMin = d
			iMin = i
		}
	}
	return iMin, dMin
}

func (s *State) KMeansQuantizingFilter(data []ClusteredColour, mean []rl.Color) {
	// initial assignment
	for i, p := range data {
		cMin, _ := Nearest(p, mean)
		data[i].cluster = cMin
	}
	mLen := make([]int, len(mean))
	for {
		// update means
		for i := range mean {
			mean[i] = rl.Color{}
			mLen[i] = 0
		}
		for _, p := range data {
			mean[p.cluster].R += p.colour.R
			mean[p.cluster].R += p.colour.G
			mean[p.cluster].R += p.colour.B
			mLen[p.cluster]++
		}
		for i := range mean {
			inv := 1 / float64(mLen[i])
			mean[i].R = uint8(inv * float64(mean[i].R))
			mean[i].G = uint8(inv * float64(mean[i].G))
			mean[i].B = uint8(inv * float64(mean[i].B))
		}
		// make new assignments, count changes
		var changes int
		for i, p := range data {
			if cMin, _ := Nearest(p, mean); cMin != p.cluster {
				changes++
				data[i].cluster = cMin
			}
		}
		if changes == 0 {
			return
		}
	}
}
func (s *State) QuantizingFilter() {
	// maybe preserve current code and make a "simple" and "accurate" mode
	// this is an NP-hard algorithm as we need to ca
	// We need to do k means clustering on the pixel space
	// then set each colour in the local space to the mean
	// value of the space... I think

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
	if s.Filters.ChannelAdjustmentEnabled {
		DebugLog("Channel adjustment applied")
		for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
			s.WorkingImage.Pix[i+0] = uint8(float32(s.WorkingImage.Pix[i+0]) * s.Filters.ChannelAdjustment[0])
			s.WorkingImage.Pix[i+1] = uint8(float32(s.WorkingImage.Pix[i+1]) * s.Filters.ChannelAdjustment[1])
			s.WorkingImage.Pix[i+2] = uint8(float32(s.WorkingImage.Pix[i+2]) * s.Filters.ChannelAdjustment[2])

		}
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
		IsGrayscaleEnabled:       false,
		IsDitheringEnabled:       false,
		DitheringLevel:           4,
		IsQuantizingEnabled:      false,
		QuantizingBands:          50,
		ChannelAdjustmentEnabled: false,
		ChannelAdjustment:        [3]float32{1.0, 1.0, 1.0},
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
