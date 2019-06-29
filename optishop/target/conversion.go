package target

import "github.com/unixpickle/optishop-server/optishop"

// Convert a MapInfo into a generic layout containing
// portals and basic zone information.
//
// This layout does not include, obstacles, aisle zones,
// or a boundary.
func MapInfoToLayout(m *MapInfo) *optishop.Layout {
	layout := &optishop.Layout{Floors: make([]*optishop.Floor, len(m.FloorMapDetails))}
	markerToPortal := map[*MapMarker]*optishop.Portal{}
	for i, floorMap := range m.FloorMapDetails {
		floor := &optishop.Floor{Name: floorMap.FloorID}
		for _, marker := range floorMap.MapMarkers {
			if marker.Portal() {
				continue
			}
			floor.Zones = append(floor.Zones, &optishop.Zone{
				Name: marker.Name,
				Location: optishop.Point{
					X: marker.Location.CenterX,
					Y: marker.Location.CenterY,
				},
				Entrance: marker.Name == "entrance",
				Checkout: marker.Name == "checkout",
			})
		}
		for _, portalMarker := range floorMap.Portals() {
			portal := &optishop.Portal{
				Location: optishop.Point{
					X: portalMarker.Location.CenterX,
					Y: portalMarker.Location.CenterY,
				},
				Type: portalTypeForMarker(portalMarker),
				ID:   len(markerToPortal),
			}
			markerToPortal[portalMarker] = portal
			floor.Portals = append(floor.Portals, portal)
		}
		if i > 0 {
			matches := MatchFloorPortals(m.FloorMapDetails[i-1], floorMap)
			for _, match := range matches {
				p1 := markerToPortal[match[0]]
				p2 := markerToPortal[match[1]]
				p1.Destinations = append(p1.Destinations, p2.ID)
				p2.Destinations = append(p2.Destinations, p1.ID)
			}
		}
		layout.Floors[i] = floor
	}
	return layout
}

// AddFloorDetails adds FloorDetails to a Floor.
func AddFloorDetails(src *FloorDetails, dst *optishop.Floor) {
	dst.Bounds = src.Bounds
	dst.Obstacles = src.Obstacles
	for name, loc := range src.Aisles {
		dst.Zones = append(dst.Zones, &optishop.Zone{
			Name:     name,
			Location: loc,
		})
	}
}

func portalTypeForMarker(m *MapMarker) optishop.PortalType {
	if m.Name == "elevators" {
		return optishop.Elevator
	} else if m.Name == "escalators" {
		return optishop.Escalator
	} else {
		return ""
	}
}
