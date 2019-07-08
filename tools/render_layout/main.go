// Command render_layout renders a Layout from its JSON
// representation.
// The JSON is read from standard input, and an SVG is
// written to standard output.
package main

import (
	"encoding/json"
	"os"

	"github.com/ajstarks/svgo/float"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/visualize"
)

func main() {
	var layout optishop.Layout
	essentials.Must(json.NewDecoder(os.Stdin).Decode(&layout))

	width, height, _ := visualize.MultiFloorGeometry(&layout)

	canvas := svg.New(os.Stdout)
	canvas.Start(width, height)

	visualize.DrawFloors(canvas, &layout)

	canvas.End()
}
