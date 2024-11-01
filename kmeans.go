package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"math"
)

type ClusteredColour struct {
	colour  rl.Color
	cluster int
}

func Dist(args ...uint8) float64 {
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
