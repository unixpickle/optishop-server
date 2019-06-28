// Command render_connector renders the output of a
// Connector on a single floor.
// The JSON for the layout is read from standard input,
// and an SVG is written to standard output.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajstarks/svgo/float"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/tools"
)

const MarginFrac = 0.1

func main() {
	if len(os.Args) != 3 {
		essentials.Die("Usage: render_connector <start_zone> <end_zone>")
	}

	startZone := os.Args[1]
	endZone := os.Args[2]

	var layout optishop.Layout
	essentials.Must(json.NewDecoder(os.Stdin).Decode(&layout))

	var startPoint, endPoint optishop.Point
	var floor *optishop.Floor
	for _, f := range layout.Floors {
		z1 := f.Zone(startZone)
		z2 := f.Zone(endZone)
		if z1 == nil && z2 == nil {
			continue
		}
		if z1 == nil || z2 == nil {
			essentials.Die("zones must be on same floor")
		}
		startPoint = z1.Location
		endPoint = z2.Location
		floor = f
		break
	}
	if floor == nil {
		essentials.Die("zones not found")
	}

	path := optishop.NewConnector(floor).Connect(startPoint, endPoint)
	if path == nil {
		essentials.Die("no path could be found")
	}

	x, y, width, height := floor.Bounds.Bounds()

	canvas := svg.New(os.Stdout)
	canvas.Start(width, height)

	tools.DrawPolygon(canvas, floor.Bounds, -x, -y, "fill: white")
	for _, obstacle := range floor.Obstacles {
		tools.DrawPolygon(canvas, obstacle, -x, -y, "fill: #d5d5d5")
	}

	var pathX, pathY []float64
	for _, p := range path {
		pathX = append(pathX, p.X-x)
		pathY = append(pathY, p.Y-y)
	}
	lineWidth := fmt.Sprintf("%.3f", width*0.005)
	canvas.Polyline(pathX, pathY, "stroke-width: "+lineWidth+"px; stroke: #2ed941; fill: none")

	canvas.End()
}
