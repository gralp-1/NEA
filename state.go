package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
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

//state object
type State struct {
	ImageLoaded bool
	ImagePath   string

	// We have current image which never changes and shown image is the one that is shown on the screen and edited
	OrigImage    image.RGBA // NOTE: making this a pointer caused a big pass by reference / pass by value bug meaning that filters couldn't be unapplied'
	WorkingImage image.RGBA
	ShownImage   *rl.Image
	ImagePalette []rl.Color
	
	// Window data
	FilterWindow   FilterOrderWindow
	PaletteWindow  PaletteWindow
	HelpWindow     HelpWindow
	SaveLoadWindow SaveLoadWindow
	SettingsWindow SettingsWindow
	
	// Histogram data
	RedHistogram   [256]int
	BlueHistogram  [256]int
	GreenHistogram [256]int

	// This is the texture that
	CurrentTexture   rl.Texture2D
	BackgroundColour rl.Color

	// UI
	Filters Filters
	
	// Config data
	Config       Config
	LanguageData [LanguageCount]map[string]string
}
// filters object
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

func (s *State) GetFiltersListViewString() string {
	return strings.Join(MapOut(s.Filters.Order[:], Translate), ";")
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

// clamp values of arbitrary types
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
	bounds := s.WorkingImage.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	// for each iteration
	for i := range s.Filters.BoxBlurIterations {
		_ = i
		// for each pixel
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				// if we're on an edge, skip
				if x < 1 || y < 1 || x+1 == w || y+1 == h {
					continue
				}
				// get the 3x3 surrounding pixels
				pixels := make([]color.Color, 9)
				pixels[0] = s.WorkingImage.At(x-1, y-1) // top left
				pixels[1] = s.WorkingImage.At(x, y-1)   // top middle
				pixels[2] = s.WorkingImage.At(x+1, y-1) // top right

				pixels[3] = s.WorkingImage.At(x-1, y) // middle left
				pixels[4] = s.WorkingImage.At(x, y)   // centre
				pixels[5] = s.WorkingImage.At(x+1, y) // middle right

				pixels[6] = s.WorkingImage.At(x-1, y+1)// bottom left  
				pixels[7] = s.WorkingImage.At(x, y+1)  // bottom middle
				pixels[8] = s.WorkingImage.At(x+1, y+1)// bottom right 
					
				// in the 3x3 kernel get the mean of red green and blue
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
				// set the pixel to the mean of the surrounding pixels and itself
				s.WorkingImage.Set(x, y, color.RGBA{R: uint8(rSum), G: uint8(gSum), B: uint8(bSum), A: uint8(aSum)})
			}
		}
	}
}

func (s *State) RefreshImage() {
	// CONSTRUCT IMAGE PALETTE MAP
	s.ApplyFilters()                                             // up to 145ms
	s.CurrentTexture = rl.LoadTextureFromImage(state.ShownImage) // >1ms
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
	s.CurrentTexture = rl.LoadTextureFromImage(s.ShownImage)
	rl.SetWindowSize(int(state.ShownImage.Width+400), int(state.ShownImage.Height))
}

func (s *State) GrayscaleFilter() {
	DebugLog("Grayscale filter applied")
	// for each pixel 
	for i := 0; i < len(s.WorkingImage.Pix); i += 4 {
		// set the r, g and b to the mean value of r, g and b
		mean := uint8((int(s.WorkingImage.Pix[i]) + int(s.WorkingImage.Pix[i+1]) + int(s.WorkingImage.Pix[i+2])) / 3)
		s.WorkingImage.Pix[i+0] = mean
		s.WorkingImage.Pix[i+1] = mean
		s.WorkingImage.Pix[i+2] = mean
	}
}

func (s *State) QuantizingFilter() {
	DebugLog("Quantizing filter applied")

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


// Floyd-Steinburg dithering
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
	// open the language file
	content, err := os.ReadFile("./resources/langs.json")
	if err != nil {
		FatalLogf("Couldn't read language data: %v", err.Error())
	}
	// deserialize it into the lanugage data map
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

}

func (s *State) SaveImage() {
	extension := s.Config.GetActiveFileFormat().String()
	InfoLogf("Saving as output.%s", extension)
	// create a file with the current format's extension
	f, err := os.Create("./output." + extension)
	if err != nil {
		FatalLogf("Couldn't create file: %v", err.Error())
	}
	// write the image to the file
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


// Close the application
func (s *State) Close() {
	// save on exit
	state.SaveImage()
	// close the window
	rl.CloseWindow()
	// successfully exit the process
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
		gui.LoadStyleDefault() // the default theme is a light theme
	case ThemeDark:
		for _, control := range DarkThemeData {
			gui.SetStyle(int32(control[0]), int32(control[1]), int64(control[2]))
		}
	}
	s.LoadFonts()
	s.ChangeFont()
}
// global font map
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
