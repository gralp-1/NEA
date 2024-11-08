package main

import (
	"time"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type FilterOrderWindow struct {
	Showing        bool
	Anchor         rl.Vector2
	ScrollIndex    int32
	Active         int32
	InteractedWith time.Time
}

func (f *FilterOrderWindow) getRect() rl.Rectangle {
	return rl.NewRectangle(f.Anchor.X, f.Anchor.Y, 300, 260)
}
func (f *FilterOrderWindow) Draw() {
	f.Showing = !gui.WindowBox(f.getRect(), "Filter Order Configuration")
	stashStyle := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+30, 100, 10), "Applied first")
	f.Active = gui.ListView(
		rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+50, 100, 180),
		state.GetFiltersListViewString(),
		&f.ScrollIndex,
		f.Active,
	)
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+240, 100, 10), "Applied last")
	// Promote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+30, 100, 50), "Promote Selected") {
		f.Promote()
	}
	// Demote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+80, 100, 50), "Demote Selected") {
		f.Demote()
	}
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, stashStyle)

}
func (f *FilterOrderWindow) Promote() {
	if f.Active == 0 {
		DebugLog("Attempted to promote first index")
	} else {
		state.Filters.Order[f.Active], state.Filters.Order[f.Active-1] = state.Filters.Order[f.Active-1], state.Filters.Order[f.Active]
		f.Active--
		state.RefreshImage()
	}
}
func (f *FilterOrderWindow) Demote() {
	if f.Active == FilterCount-1 {
		DebugLog("Attempted to demote last index")
	} else {
		state.Filters.Order[f.Active], state.Filters.Order[f.Active+1] = state.Filters.Order[f.Active+1], state.Filters.Order[f.Active]
		f.Active++
		state.RefreshImage()
	}
}
