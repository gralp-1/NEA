package main

import (
	"slices"
	"testing"
)

func TestClamp(t *testing.T) {
	t.Run("Test Negative", func(t *testing.T) {
		if ClampByte(-100) != 0 {
			t.Fail()
		}
	})
	t.Run("Test Above", func(t *testing.T) {
		if ClampByte(999) != 255 {
			t.Fail()
		}
	})
	t.Run("Test in range", func(t *testing.T) {
		if ClampByte(120) != 120 {
			t.Fail()
		}
	})
	t.Run("Test in upper bound", func(t *testing.T) {
		if ClampByte(255) != 255 {
			t.Fail()
		}
	})
	t.Run("Test in lower bound", func(t *testing.T) {
		if ClampByte(0) != 0 {
			t.Fail()
		}
	})
	t.Run("Test out upper bound", func(t *testing.T) {
		if ClampByte(256) != 255 {
			t.Fail()
		}
	})
	t.Run("Test out lower bound", func(t *testing.T) {
		if ClampByte(-1) != 0 {
			t.Fail()
		}
	})
}

func TestRemoveDuplicates(t *testing.T) {
	t.Run("Test all duplicates", func(t *testing.T) {
		lhs := []int{1}
		rhs := []int{1, 1, 1, 1, 1, 1}
		if !slices.Equal(lhs, removeDuplicates(rhs)) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
	t.Run("Test no duplicates", func(t *testing.T) {
		lhs := []int{1, 2, 3, 4, 5, 6}
		rhs := []int{1, 2, 3, 4, 5, 6}
		if !slices.Equal([]int{1, 2, 3, 4, 5, 6}, removeDuplicates([]int{1, 2, 3, 4, 5, 6})) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
	t.Run("Test empty", func(t *testing.T) {
		lhs := []int{}
		rhs := []int{}
		if !slices.Equal(lhs, removeDuplicates(rhs)) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
	t.Run("Test single item", func(t *testing.T) {
		lhs := []int{1}
		rhs := []int{1}
		if !slices.Equal([]int{1}, removeDuplicates([]int{1})) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
	t.Run("Test one duplicate", func(t *testing.T) {
		lhs := []int{1, 2, 3, 4, 5, 6}
		rhs := []int{1, 1, 2, 3, 4, 5, 6}
		if !slices.Equal([]int{1, 2, 3, 4, 5, 6}, removeDuplicates([]int{1, 1, 2, 3, 4, 5, 6})) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
	t.Run("Test each element duplicated", func(t *testing.T) {
		lhs := []int{1, 2, 3, 4, 5, 6}
		rhs := []int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6}
		if !slices.Equal([]int{1, 2, 3, 4, 5, 6}, removeDuplicates([]int{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6})) {
			t.Errorf("LHS: %v\nRHS: %v", lhs, rhs)
		}
	})
}
func TestQuantizeValue(t *testing.T) {
	t.Run("Test zero bands", func(t *testing.T) {
		res := make([]uint8, 256)
		for i := 0; i < 256; i++ {
			res[i] = Quantize(0, uint8(i))
		}
		RHS := make([]uint8, 256)
		if !slices.Equal(res, RHS) {
			t.Errorf("Expected %v, got %v", RHS, res)
		}
	})
	t.Run("Test 255 bands", func(t *testing.T) {
		for i := range 255 {
			res := Quantize(255, uint8(i))
			if uint8(i) != res {
				t.Errorf("Expected %d got %d", i, res)
			}
		}
	})
	t.Run("Test 1 band", func(t *testing.T) {
		for i := range 255 {
			if Quantize(1, uint8(i)) != uint8(i) {
				t.Errorf("Expected %d got %d", i, Quantize(1, uint8(i)))
			}
		}
	})
	t.Run("Test normal case", func(t *testing.T) {
		bandCount := uint8(4)
		res := make([]uint8, 256)
		for i := 0; i < 256; i++ {
			res[i] = Quantize(bandCount, uint8(i))
		}
		removeDuplicates(res)
		expected := []uint8{0}
		if !slices.Equal(res, expected) {
			t.Errorf("Expected %v, got %v", expected, res)
		}

	})
}
