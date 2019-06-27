package optishop

// A PortalType is a way of getting from one floor to
// another in a store.
type PortalType string

const (
	Elevator  PortalType = "elevator"
	Escalator PortalType = "escalator"
)

// A Layout stores the physical layout of a store.
type Layout struct {
	Floors []Floor
}

// Portal finds the portal with the given ID, or returns
// nil if no portal was found.
func (l *Layout) Portal(id int) *Portal {
	for _, f := range l.Floors {
		if p := f.Portal(id); p != nil {
			return p
		}
	}
	return nil
}

// A Floor stores the physical layout of a single floor of
// a store.
type Floor struct {
	Name    string
	Zones   []*Zone
	Portals []*Portal
}

// Portal finds the portal with the given ID, or returns
// nil if no portal was found.
func (f *Floor) Portal(id int) *Portal {
	for _, p := range f.Portals {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// A Zone is an arbitrary location in a store.
type Zone struct {
	// If not empty, may be an aisle name or a department
	// name, etc.
	Name string

	// Spacial coordinates within the floor.
	//
	// All coordinates are relative to the floor, and all
	// distances are to scale within a floor.
	X float64
	Y float64
}

// A Portal is a means by which a customer can get from
// one Floor of a Layout to another.
type Portal struct {
	// See Zone for documentation on coordinates.
	X float64
	Y float64

	Type PortalType

	// An identifier for the portal that is unique to the
	// Layout containing the portal.
	ID int

	// All of the portal IDs to which this portal leads.
	Destinations []int
}
