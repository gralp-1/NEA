package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// clamp in 0-255 range
func clamp(value int) int {
	if value < 0 {
		return 0
	} else if value > 255 {
		return 255
	}
	return value
}
func distributeError(img *image.RGBA, x, y int, errR, errG, errB uint8, factor float64) {
	bounds := img.Bounds()
	xBoundInImage := x >= bounds.Min.X && x < bounds.Max.X
	yBoundInImage := y >= bounds.Min.Y && y < bounds.Max.Y
	if xBoundInImage && yBoundInImage {
		currR, currG, currB, _ := img.At(x, y).RGBA()

		currR /= 257
		currG /= 257
		currB /= 257

		newR := clamp(int(currR) + int(float64(errR)*factor))
		newG := clamp(int(currG) + int(float64(errG)*factor))
		newB := clamp(int(currB) + int(float64(errB)*factor))
		img.SetRGBA(x, y, color.RGBA{R: uint8(newR), G: uint8(newG), B: uint8(newB), A: 255})
	}
}

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

type Integer interface {
	uint8 | uint16 | uint32 | uint64 | uintptr |
		int8 | int16 | int32 | int64
}

func chunk[T Integer](slice []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

func PixSliceToColourSlice(pix []uint8) []rl.Color {
	if len(pix)%4 != 0 {
		panic("Passed in pix slice with length not a multiple of 4")
	}
	pixels := chunk(pix, 4)
	res := make([]rl.Color, len(pix)/4)
	for i, p := range pixels {
		res[i/4] = rl.NewColor(p[0], p[1], p[2], p[3])
	}
	return res
}
func QuantizeValue(bandCount, v uint8) uint8 {
	// TODO: make some adjustable curve??
	bandWidth := uint8(math.Floor(256 / float64(bandCount)))
	for i := range bandCount + 1 {
		if (bandCount-i)*bandWidth < v {
			return bandWidth * (bandCount - i)
		}
	}
	return 0
}

func Translate(in string) string {
	res := state.LanguageData[state.Config.Language][in]
	if res == "" {
		languageName := ""
		switch state.Config.Language {
		case English:
			languageName = "English"
		case German:
			languageName = "German"
		default:
			FatalLogf("Language %v has no name for settings menu", state.Config.Language)
		}
		return fmt.Sprintf("%v not translated to %v", in, languageName)
	}
	return res
}

func FilterOut[T any](vals []T, f func(T) bool) []T {
	res := make([]T, len(vals))
	for _, v := range vals {
		if f(v) {
			res = append(res, v)
		}
	}
	return res
}

func MapOut[T any](vals []T, f func(T) T) []T {
	res := make([]T, len(vals))
	for i, v := range vals {
		res[i] = f(v)
	}
	return res
}
