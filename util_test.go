package main

import (
	"slices"
	"testing"
)

func TestClamp(t *testing.T) {
	t.Run("Test Negative", func(t *testing.T) {
		if clamp(-100) != 0 {
			t.Fail()
		}
	})
	t.Run("Test Above", func(t *testing.T) {
		if clamp(999) != 255 {
			t.Fail()
		}
	})
	t.Run("Test in range", func(t *testing.T) {
		if clamp(120) != 120 {
			t.Fail()
		}
	})
	t.Run("Test in upper bound", func(t *testing.T) {
		if clamp(255) != 255 {
			t.Fail()
		}
	})
	t.Run("Test in lower bound", func(t *testing.T) {
		if clamp(0) != 0 {
			t.Fail()
		}
	})
	t.Run("Test out upper bound", func(t *testing.T) {
		if clamp(256) != 255 {
			t.Fail()
		}
	})
	t.Run("Test out lower bound", func(t *testing.T) {
		if clamp(-1) != 0 {
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
		if !slices.Equal([]int{}, removeDuplicates([]int{})) {
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
func TestStack(t *testing.T) {
	t.Run("Test push", func(t *testing.T) {
		s := NewStack[int]()
		s.Push(10)
		s.Push(20)
		s.Push(30)
		s.Push(40)
		if !slices.Equal([]int{10, 20, 30, 40}, s.items) {
			t.Errorf("Expected [10, 20, 30, 40] and got %v", s.items)
		}
	})
	t.Run("Test len and pop", func(t *testing.T) {
		s := NewStack[int]()
		s.Push(10)
		s.Push(20)
		if s.Len() != 2 {
			t.Fail()
		}
		s.Pop()
		if s.Len() != 1 {
			t.Fail()
		}
		s.Push(20)
		if s.Len() != 2 {
			t.Fail()
		}
	})
	t.Run("Test peek", func(t *testing.T) {
		s := NewStack[int]()
		s.Push(10)
		if s.Peek() != 10 {
			t.Fail()
		}
		s.Push(20)
		if s.Peek() != 20 {
			t.Fail()
		}
		if s.Len() != 2 { // this means peek is probably removing elements
			t.Fail()
		}
	})
}
