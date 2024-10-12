package main

import (
	"cmp"
	"slices"
)

type Set[T cmp.Ordered] struct {
	List []T
}

func (s *Set[T]) Insert(item T) {
	if slices.Contains[[]T, T](s.List, item) {
		return
	}
	s.List = append(s.List, item)
	slices.Sort(s.List)
}
func (s *Set[T]) Remove(item T) {
	if !slices.Contains[[]T, T](s.List, item) {
		return
	}
	slices.DeleteFunc(s.List, func(E T) bool { return E == item })
}
func NewSet[T cmp.Ordered](initialItems ...T) Set[T] {
	res := make([]T, len(initialItems))
	for _, v := range initialItems {
		if slices.Contains[[]T, T](res, v) {
			continue
		}
		res = append(res, v)
	}
	return Set[T]{List: res}
}
