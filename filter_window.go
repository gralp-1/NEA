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
// Draw the filter order window
func (f *FilterOrderWindow) Draw() {
	// Draw the window box
	f.Showing = !gui.WindowBox(f.getRect(), Translate("window.filter.title"))
	// Store all the prexisting style data before we align the text to centre for the order list
	stashStyle := gui.GetStyle(gui.LABEL, gui.TEXT_ALIGNMENT)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	// Draw the "Applied first" label
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+30, 100, 10), Translate("window.filter.appliedfirst"))
	// Draw the filter order list
	f.Active = gui.ListView(
		rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+50, 100, 180),
		state.GetFiltersListViewString(),
		&f.ScrollIndex,
		f.Active,
	)
	// Draw the "Applied last" label
	gui.Label(rl.NewRectangle(f.Anchor.X+10, f.Anchor.Y+240, 100, 10), Translate("window.filter.appliedlast"))
	// Promote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+30, 100, 50), Translate("window.filter.promote")) {
		f.Promote()
	}
	// Demote button
	if gui.Button(rl.NewRectangle(f.Anchor.X+150, f.Anchor.Y+80, 100, 50), Translate("window.filter.demote")) {
		f.Demote()
	}
	// Put the style back
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, stashStyle)

}
func (f *FilterOrderWindow) Promote() {
	// if the selected is the top dont't do anything
	if f.Active == 0 {
		DebugLog("Attempted to promote first index")
	} else {
		// else swap it with the one below
		state.Filters.Order[f.Active], state.Filters.Order[f.Active-1] = state.Filters.Order[f.Active-1], state.Filters.Order[f.Active]
		// and move the active down
		f.Active--
	}
}
func (f *FilterOrderWindow) Demote() {
	// if the selected is the bottom dont't do anything
	if f.Active == FilterCount-1 {
		DebugLog("Attempted to demote last index")
	} else {
		// else swap it with the one above
		state.Filters.Order[f.Active], state.Filters.Order[f.Active+1] = state.Filters.Order[f.Active+1], state.Filters.Order[f.Active]
		// and move the active one up
		f.Active++
	}
}
