package target

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
)

// FloorDetails stores very specific information about a
// certain floor of a store.
type FloorDetails struct {
	Aisles    map[string]optishop.Point
	Obstacles []optishop.Polygon
	Bounds    optishop.Polygon
}

// GetFloorDetails looks up the floor details from the map
// of a specific floor of a specific store.
func GetFloorDetails(storeID, floorID string) (*FloorDetails, error) {
	url := "https://prod.tgtneptune.com/v1/stores/" + storeID + "/maps/svgs/floors/" + floorID
	data, err := GetRequest(url)
	if err != nil {
		return nil, errors.Wrap(err, "get floor details")
	}
	if details, err := parseFloorDetails(data); err != nil {
		return nil, errors.Wrap(err, "get floor details")
	} else {
		return details, nil
	}
}

func parseFloorDetails(data []byte) (*FloorDetails, error) {
	parsed, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	content, ok := scrape.Find(parsed, scrape.ById("content"))
	if !ok {
		return nil, errors.New("missing 'content' group")
	}
	transform, err := parseCoordTransform(scrape.Attr(content, "transform"))
	if err != nil {
		return nil, err
	}

	result := &FloorDetails{Aisles: map[string]optishop.Point{}}

	wallShapes, ok := scrape.Find(parsed, scrape.ById("Wall-Shapes"))
	if !ok {
		return nil, errors.New("missing 'Wall-Shapes' group")
	}
	paths := findTag(wallShapes, "path")
	if len(paths) != 1 {
		return nil, errors.New("expected exactly one wall shape")
	}
	result.Bounds, err = pathPolygon(paths[0], transform)
	if err != nil {
		return nil, err
	}

	aisleShapes, ok := scrape.Find(parsed, scrape.ById("Aisle-Shapes"))
	if !ok {
		return nil, errors.New("missing 'Aisle-Shapes' group")
	}
	for _, path := range findTag(aisleShapes, "path") {
		poly, err := pathPolygon(path, transform)
		if err != nil {
			return nil, err
		}
		result.Obstacles = append(result.Obstacles, poly)
	}

	aisleNames, ok := scrape.Find(parsed, scrape.ById("Aisle-Names"))
	if !ok {
		return nil, errors.New("missing 'Aisle-Names' group")
	}
	for _, text := range findTag(aisleNames, "text") {
		xText := scrape.Attr(text, "x")
		yText := scrape.Attr(text, "y")
		x, err := strconv.ParseFloat(xText, 64)
		if err != nil {
			return nil, err
		}
		y, err := strconv.ParseFloat(yText, 64)
		if err != nil {
			return nil, err
		}
		result.Aisles[strings.TrimSpace(scrape.Text(text))] = transform.Apply(optishop.Point{
			X: x,
			Y: y,
		})
	}

	return result, nil
}

func findTag(elem *html.Node, tag string) []*html.Node {
	return scrape.FindAll(elem, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	})
}

// A coordTransform is a 2-D transformation matrix in the
// order defined in the SVG spec.
type coordTransform [6]float64

func parseCoordTransform(transform string) (*coordTransform, error) {
	if !strings.HasPrefix(transform, "matrix(") || !strings.HasSuffix(transform, ")") {
		return nil, errors.New("unsupported transform: " + transform)
	}
	comps := strings.Fields(transform[7 : len(transform)-1])
	if len(comps) != 6 {
		return nil, errors.New("unexpected term count in transform: " + transform)
	}
	var res coordTransform
	for i, comp := range comps {
		x, err := strconv.ParseFloat(comp, 64)
		if err != nil {
			return nil, errors.New("unexpected term in transform: " + transform)
		}
		res[i] = x
	}
	return &res, nil
}

func (c *coordTransform) Apply(p optishop.Point) optishop.Point {
	return optishop.Point{
		X: p.X*c[0] + p.Y*c[2] + c[4],
		Y: p.X*c[1] + p.Y*c[3] + c[5],
	}
}

func pathPolygon(elem *html.Node, c *coordTransform) (optishop.Polygon, error) {
	polygon := optishop.Polygon{}
	data := scrape.Attr(elem, "d")
	fields := strings.Fields(data)
	if len(fields) == 0 {
		return nil, errors.New("empty path")
	}
	if len(fields)%2 != 1 || fields[len(fields)-1] != "Z" {
		return nil, errors.New("expected even # of fields followed by Z")
	}
	for i := 0; i < len(fields)-2; i += 2 {
		cmd1 := fields[i]
		cmd2 := fields[i+1]
		if i == 0 {
			if !strings.HasPrefix(cmd1, "M") {
				return nil, errors.New("expected move command")
			}
		} else if !strings.HasPrefix(cmd1, "L") {
			return nil, errors.New("expected line command")
		}
		val1, err := strconv.ParseFloat(cmd1[1:], 64)
		if err != nil {
			return nil, err
		}
		val2, err := strconv.ParseFloat(cmd2, 64)
		if err != nil {
			return nil, err
		}
		polygon = append(polygon, c.Apply(optishop.Point{X: val1, Y: val2}))
	}
	return polygon, nil
}
