package main

import (
	"fmt"
	"image"
	"image/color"

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

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{
		items: make([]T, 0),
	}
}

func (s *Stack[T]) Len() uint {
	return uint(len(s.items))
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}
func (s *Stack[T]) Peek() T {
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) Pop() {
	if s.IsEmpty() {
		return
	}
	s.items = s.items[:len(s.items)-1]
}

func (s *Stack[T]) Top() (T, error) {
	var out T // basically nil for [T any]
	if s.IsEmpty() {
		return out, fmt.Errorf("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

func (s *Stack[T]) IsEmpty() bool {
	if len(s.items) == 0 {
		return true
	}
	return false
}

func (s *Stack[T]) Print() {
	for _, item := range s.items {
		fmt.Print(item, " ")
	}
	fmt.Println()
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
