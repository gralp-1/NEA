package main

import (
	"fmt"
	"image"
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ClampByte in 0-255 range
func ClampByte(value int) int {
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

		newR := ClampByte(int(currR) + int(float64(errR)*factor))
		newG := ClampByte(int(currG) + int(float64(errG)*factor))
		newB := ClampByte(int(currB) + int(float64(errB)*factor))
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

// QuantizeValue quantizes a value into a bucketCount number of buckets
// exceptions: bandCount = 0 -> return 0
// bandCount = 1 -> return value
func Quantize(bandCount, v uint8) uint8 {
	if bandCount == 1 || bandCount == 0 {
		return v
	}
	// quantize a function into bandCount buckets
	bucketSize := uint8(math.Trunc(float64(255) / float64(bandCount+1)))
	return uint8(math.Trunc(float64(v)/float64(bucketSize)) * float64(bucketSize))
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
