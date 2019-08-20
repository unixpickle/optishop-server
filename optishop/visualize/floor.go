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
func DrawFloor(canvas *svg.SVG, floor *optishop.Floor, xOff, yOff, fontSize float64) {
	DrawFloorPolygons(canvas, floor, xOff, yOff)
	DrawFloorLabels(canvas, floor, xOff, yOff, fontSize)
}

// DrawFloorPolygons draws all of the objects on a floor
// except for the labels.
func DrawFloorPolygons(canvas *svg.SVG, floor *optishop.Floor, xOff, yOff float64) {
	DrawPolygon(canvas, floor.Bounds, xOff, yOff, "fill: white")
	for _, nonPref := range floor.NonPreferred {
		if nonPref.Visible {
			DrawPolygon(canvas, nonPref.Bounds, xOff, yOff, "fill: #f0f0f0")
		}
	}
	for _, obstacle := range floor.Obstacles {
		DrawPolygon(canvas, obstacle, xOff, yOff, "fill: #d5d5d5")
	}
}

// DrawFloorLabels draws all of the zone labels.
func DrawFloorLabels(canvas *svg.SVG, floor *optishop.Floor, xOff, yOff, fontSize float64) {
	DrawZoneLabels(canvas, floor.Zones, xOff, yOff, fontSize)
}

// DrawZoneLabels draws the labels for specified zones.
func DrawZoneLabels(canvas *svg.SVG, zones []*optishop.Zone, xOff, yOff, fontSize float64) {
	for _, zone := range zones {
		fs := fontSize
		if zone.Specific {
			fs *= SpecificLabelSizeFrac
		}
		sizeStr := fmt.Sprintf("%.3fpx", fs)
		canvas.Text(zone.Location.X+xOff, zone.Location.Y+yOff, zone.Name,
			"text-anchor: middle; dominant-baseline: middle; font-size: "+sizeStr)
	}
}
