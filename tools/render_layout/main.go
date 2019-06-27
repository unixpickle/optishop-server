// Command render_layout renders a Layout from its JSON
// representation.
// The JSON is read from standard input, and an SVG is
// written to standard output.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/ajstarks/svgo/float"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
)

const MarginFrac = 0.1

func main() {
	var layout optishop.Layout
	essentials.Must(json.NewDecoder(os.Stdin).Decode(&layout))

	width, height, margin := computeGeometries(layout)
	canvas := svg.New(os.Stdout)
	canvas.Start(width, height)

	fontSize := fmt.Sprintf("%.3fpx", width/150)

	destY := margin
	for _, floor := range layout.Floors {
		x, y, curWidth, height := floor.Bounds.Bounds()
		destX := (width - curWidth) / 2

		drawPolygon(canvas, floor.Bounds, destX-x, destY-y, "fill: white")
		for _, obstacle := range floor.Obstacles {
			drawPolygon(canvas, obstacle, destX-x, destY-y, "fill: #d5d5d5")
		}
		for _, zone := range floor.Zones {
			canvas.Text(zone.Location.X+destX-x, zone.Location.Y+destY-y, zone.Name,
				"text-anchor: middle; dominant-baseline: middle; font-size: "+fontSize)
		}

		destY += height + margin
	}

	canvas.End()
}

func computeGeometries(layout optishop.Layout) (width, height, margin float64) {
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

func drawPolygon(canvas *svg.SVG, poly optishop.Polygon, xOff, yOff float64, style string) {
	var xs, ys []float64
	for _, p := range poly {
		xs = append(xs, p.X+xOff)
		ys = append(ys, p.Y+yOff)
	}
	canvas.Polygon(xs, ys, style)
}
