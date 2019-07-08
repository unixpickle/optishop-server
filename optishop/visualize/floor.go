package visualize

import (
	"fmt"

	svg "github.com/ajstarks/svgo/float"
	"github.com/unixpickle/optishop-server/optishop"
)

// DrawPolygon fills in a polygon on the canvas.
//
// All points are offset by (xOff, yOff).
//
// The CSS style string is used to style the polygon.
func DrawPolygon(canvas *svg.SVG, poly optishop.Polygon, xOff, yOff float64, style string) {
	var xs, ys []float64
	for _, p := range poly.Dedup() {
		xs = append(xs, p.X+xOff)
		ys = append(ys, p.Y+yOff)
	}
	canvas.Polygon(xs, ys, style)
}

// DrawFloor draws all of the objects on a floor.
func DrawFloor(canvas *svg.SVG, floor *optishop.Floor, fontSize, xOff, yOff float64) {
	DrawPolygon(canvas, floor.Bounds, xOff, yOff, "fill: white")
	for _, nonPref := range floor.NonPreferred {
		DrawPolygon(canvas, nonPref, xOff, yOff, "fill: #f0f0f0")
	}
	for _, obstacle := range floor.Obstacles {
		DrawPolygon(canvas, obstacle, xOff, yOff, "fill: #d5d5d5")
	}
	sizeStr := fmt.Sprintf("%.3fpx", fontSize)
	for _, zone := range floor.Zones {
		canvas.Text(zone.Location.X+xOff, zone.Location.Y+yOff, zone.Name,
			"text-anchor: middle; dominant-baseline: middle; font-size: "+sizeStr)
	}
}
