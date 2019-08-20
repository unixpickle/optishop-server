package visualize

import (
	"fmt"
	"math"

	svg "github.com/ajstarks/svgo/float"
	"github.com/unixpickle/optishop-server/optishop"
)

// MarginFrac controls how big the margins of a rendering
// are with respect to the size of the layout.
const MarginFrac = 0.1

// FontSizeFrac controls how big the department labels are
// with respect to the size of the layout.
const FontSizeFrac = 1.0 / 150.0

// SpecificLabelSizeFrac controls how big the aisle labels are with
// respect to the size of the department labels.
const SpecificLabelSizeFrac = 0.5

// PathFrac controls how thick paths are with respect to
// the size of the layout.
const PathFrac = 0.002

// MultiFloorGeometry computes the total width and height
// of a rendered layout, when all the floors are rendered
// on the same image.
//
// The sizes include room for margins, and the margin is
// returned as well for convenience.
func MultiFloorGeometry(layout *optishop.Layout) (width, height, margin float64) {
	var maxWidth float64
	var totalHeight float64
	for _, floor := range layout.Floors {
		_, _, width, height := floor.Bounds.Bounds()
		maxWidth = math.Max(maxWidth, width)
		totalHeight += height
	}
	margin = MarginFrac * maxWidth
	height = totalHeight + margin*float64(len(layout.Floors)+1)
	width = maxWidth + margin*2
	return
}

// MultiFloorLoop runs a function f for each floor,
// specifying the x and y offset for that floor.
func MultiFloorLoop(layout *optishop.Layout, f func(f *optishop.Floor, x, y float64)) {
	width, _, margin := MultiFloorGeometry(layout)
	destY := margin
	for _, floor := range layout.Floors {
		x, y, curWidth, height := floor.Bounds.Bounds()
		destX := (width - curWidth) / 2
		f(floor, destX-x, destY-y)
		destY += height + margin
	}
}

// DrawFloors draws every floor of a layout.
func DrawFloors(canvas *svg.SVG, layout *optishop.Layout) {
	width, _, _ := MultiFloorGeometry(layout)
	fontSize := width * FontSizeFrac
	MultiFloorLoop(layout, func(f *optishop.Floor, x, y float64) {
		DrawFloor(canvas, f, x, y, fontSize)
	})
}

// DrawFloorPath traces out a path on a multi-floor
// rendering.
func DrawFloorPath(canvas *svg.SVG, layout *optishop.Layout, path optishop.FloorPath) {
	width, _, _ := MultiFloorGeometry(layout)
	MultiFloorLoop(layout, func(f *optishop.Floor, x, y float64) {
		for _, part := range path {
			if layout.Floors[part.Floor] != f {
				continue
			}
			var pathX, pathY []float64
			for _, p := range part.Path {
				pathX = append(pathX, p.X+x)
				pathY = append(pathY, p.Y+y)
			}
			lineWidth := fmt.Sprintf("%.3f", width*PathFrac)
			canvas.Polyline(pathX, pathY,
				"stroke-width: "+lineWidth+"px; stroke: #65bcd4; fill: none")
			canvas.Circle(pathX[len(pathX)-1], pathY[len(pathY)-1], width*PathFrac*2,
				"fill: #65bcd4")
		}
	})
}
