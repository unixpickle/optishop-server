package optishop

// A PortalType is a way of getting from one floor to
// another in a store.
type PortalType string

const (
	Elevator  PortalType = "elevator"
	Escalator PortalType = "escalator"
)

// A Layout specifies the physical layout of a store.
type Layout struct {
	Floors []*Floor
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

// Zone finds the zone with the given name, or returns nil
// if no zone was found.
func (l *Layout) Zone(name string) *Zone {
	for _, f := range l.Floors {
		if z := f.Zone(name); z != nil {
			return z
		}
	}
	return nil
}

// FloorIndex finds the index of the given floor.
// Returns -1 if the floor is not found.
func (l *Layout) FloorIndex(f *Floor) int {
	for i, x := range l.Floors {
		if x == f {
			return i
		}
	}
	return -1
}

// PortalFloor gets the floor number of the given portal.
// Returns -1 if the portal is not found.
func (l *Layout) PortalFloor(p *Portal) int {
	for i, f := range l.Floors {
		for _, x := range f.Portals {
			if x == p {
				return i
			}
		}
	}
	return -1
}

// ZoneFloor gets the floor number of the given zone.
// Returns -1 if the zone is not found.
func (l *Layout) ZoneFloor(z *Zone) int {
	for i, f := range l.Floors {
		for _, x := range f.Zones {
			if x == z {
				return i
			}
		}
	}
	return -1
}

// A Floor specifies the physical layout of a single floor
// of a store.
type Floor struct {
	Name    string
	Zones   []*Zone
	Portals []*Portal

	// The containing shape of the floor. Shoppers may not
	// step outside of this shape (without exiting the
	// store).
	Bounds Polygon

	// All areas (e.g. shelves) which a shopper cannot
	// penetrate in a floor.
	Obstacles []Polygon

	// All areas (e.g. carpeted sections) which a shopper
	// shouldn't go through unless it's a destination.
	NonPreferred []*NonPreferred
}

type NonPreferred struct {
	Bounds Polygon

	// If true, render this area, perhaps as a carpeted
	// section.
	Visible bool
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

// Zone finds the zone with the given name, or returns nil
// if no zone was found.
func (f *Floor) Zone(name string) *Zone {
	for _, z := range f.Zones {
		if z.Name == name {
			return z
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
	Location Point

	// If true, this is a place to enter/exit the store.
	Entrance bool

	// If true, this is a place to pay for merchandise.
	Checkout bool

	// If true, this is a highly specific zone, e.g. an
	// aisle.
	Specific bool
}

// A Portal is a means by which a customer can get from
// one Floor of a Layout to another.
type Portal struct {
	// See Zone for documentation on coordinates.
	Location Point

	Type PortalType

	// An identifier for the portal that is unique to the
	// Layout containing the portal.
	ID int

	// All of the portal IDs to which this portal leads.
	Destinations []int
}
