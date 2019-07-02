package optishop

// A FloorConnector combines Connectors from each floor
// of a building to connect arbitrary locations in the
// building.
type FloorConnector struct {
	Connectors []Connector
	Layout     *Layout
}

// Connect finds a short FloorPath between points a and b.
func (f *FloorConnector) Connect(a, b *FloorPoint) FloorPath {
	if a.Floor == b.Floor {
		rawPath := f.Connectors[a.Floor].Connect(a.Point, b.Point)
		floorPath := FloorPath{}
		for _, loc := range rawPath {
			floorPath = append(floorPath, &FloorPathStep{
				Location: &FloorPoint{
					Point: loc,
					Floor: a.Floor,
				},
			})
		}
		return floorPath
	}
	panic("not yet implemented")
}

// A FloorPoint is a Point with extra information about
// the floor number.
type FloorPoint struct {
	// Point stores a location within the floor.
	Point

	// Floor stores the current floor index.
	Floor int
}

// A FloorPathStep is a single step in a path between two
// arbitrary places in a building.
type FloorPathStep struct {
	// Location stores the location of the user at the
	// end of this step.
	Location *FloorPoint

	// PortalType specifies the type of portal the user is
	// standing in/on at the end of this step.
	// It is the zero value of PortalType if there is no
	// portal at this step.
	PortalType PortalType
}

// A FloorPath is a set of steps that takes the user from
// one point in a store to another.
//
// The first step specifies the start position, and the
// final step specifies the end position.
type FloorPath []*FloorPathStep
