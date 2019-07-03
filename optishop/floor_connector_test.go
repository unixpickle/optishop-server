package optishop

import "testing"

func TestFloorConnector(t *testing.T) {
	layout := complexMultiFloorLayout()
	conn := NewFloorConnector(layout)

	zoneB := layout.Zone("B")
	zoneC := layout.Zone("C")
	zoneD := layout.Zone("D")

	t.Run("SameFloor", func(t *testing.T) {
		route := conn.Connect(FloorPoint{Point: zoneB.Location, Floor: 1},
			FloorPoint{Point: zoneC.Location, Floor: 1})
		if len(route) != 1 {
			t.Fatal("unexpected one step, but got", len(route))
		}
		if route[0].Path[0] != zoneB.Location {
			t.Fatal("unexpected start point")
		}
		if route[0].Path[len(route[0].Path)-1] != zoneC.Location {
			t.Fatal("unexpected end point")
		}
		if route[0].Floor != 1 {
			t.Fatal("unexpected floor")
		}
	})

	t.Run("ThreeSteps", func(t *testing.T) {
		route := conn.Connect(FloorPoint{Point: zoneB.Location, Floor: 1},
			FloorPoint{Point: zoneD.Location, Floor: 2})
		if len(route) != 3 {
			t.Fatal("expected three steps, but got", len(route))
		}
		if route[0].SourcePortal != 3 || route[0].DestPortal != 1 || route[0].Floor != 1 {
			t.Error("unexpected first step")
		}
		if route[1].SourcePortal != 1 || route[1].DestPortal != 4 || route[1].Floor != 0 {
			t.Error("unexpected second step")
		}
		for _, i := range []int{0, len(route[1].Path) - 1} {
			p := route[1].Path[i]
			if p != layout.Portal(1).Location {
				t.Error("unexpected point in second step path", p, layout.Portal(1).Location)
			}
		}
		if route[2].Floor != 2 {
			t.Error("unexpected last step")
		}
		if route[2].Path[0] != layout.Portal(4).Location {
			t.Error("unexpected start point in last step")
		}
		if route[2].Path[len(route[2].Path)-1] != zoneD.Location {
			t.Error("unexpected end point in last step")
		}
	})
}

func complexMultiFloorLayout() *Layout {
	return &Layout{
		Floors: []*Floor{
			&Floor{
				Bounds: Polygon{{0, 0}, {7, 0}, {0, 5}},
				Zones: []*Zone{
					&Zone{
						Name:     "A",
						Location: Point{1, 3},
					},
				},
				Portals: []*Portal{
					&Portal{
						Location:     Point{1, 1},
						ID:           0,
						Destinations: []int{2},
					},
					&Portal{
						Location:     Point{3, 2},
						ID:           1,
						Destinations: []int{4},
					},
				},
			},
			&Floor{
				Bounds: Polygon{{0, 0}, {3, 5}, {4, 3}, {5, 5}, {8, 0}},
				Zones: []*Zone{
					&Zone{
						Name:     "B",
						Location: Point{6, 1},
					},
					&Zone{
						Name:     "C",
						Location: Point{5, 4},
					},
				},
				Portals: []*Portal{
					&Portal{
						Location:     Point{3, 1},
						ID:           2,
						Destinations: []int{},
					},
					&Portal{
						Location:     Point{3, 4},
						ID:           3,
						Destinations: []int{1},
					},
				},
			},
			&Floor{
				Bounds: Polygon{{0, 0}, {9, 0}, {9, 4}, {0, 4}},
				Zones: []*Zone{
					&Zone{
						Name:     "D",
						Location: Point{1, 3},
					},
				},
				Portals: []*Portal{
					&Portal{
						Location:     Point{7, 1},
						ID:           4,
						Destinations: []int{1},
					},
				},
			},
		},
	}
}
