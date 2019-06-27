package optishop

// A Path is a sequence of points leading from some start
// destination to some end destination.
type Path []Point

// A Connector finds short paths from one point to another
// on a Floor, avoiding obstacles as needed.
type Connector interface {
	// Obstructed checks if a point is obstructed.
	// A point is obstructed if it is either inside of an
	// obstacle, or outside of the floor's bounds.
	Obstructed(p Point) bool

	// Unobstruct gets a point close to p which is not
	// obstructed.
	Unobstruct(p Point) Point

	// Connect finds a path connecting points a and b.
	//
	// If the start or end point is obstructed, a nearby
	// unobstructed point is used.
	//
	// If no path can be found, nil is returned.
	Connect(a, b Point) Path
}
