package tools

import (
	svg "github.com/ajstarks/svgo/float"
	"github.com/unixpickle/optishop-server/optishop"
)

func DrawPolygon(canvas *svg.SVG, poly optishop.Polygon, xOff, yOff float64, style string) {
	var xs, ys []float64
	for _, p := range poly {
		xs = append(xs, p.X+xOff)
		ys = append(ys, p.Y+yOff)
	}
	canvas.Polygon(xs, ys, style)
}
