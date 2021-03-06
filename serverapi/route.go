package serverapi

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/db"
)

// SortEntries finds the optimal route and returns the
// list entries sorted by this route.
//
// The entries are copied and their zones are replaced
// with actual pointers from store.Layout().
func SortEntries(list []*db.ListEntry, store optishop.Store,
	conn *optishop.FloorConnector) ([]*db.ListEntry, error) {
	layout := store.Layout()
	newList := make([]*db.ListEntry, len(list))
	for i, entry := range list {
		info := *entry.Info
		info.Zone = findExactZone(layout, info.Floor, info.Zone)
		if info.Zone == nil {
			return nil, fmt.Errorf("sort entries: invalid zone \"%s\" for list entry %d",
				entry.Info.Zone.Name, i)
		}
		newList[i] = &db.ListEntry{
			ID:   entry.ID,
			Info: &info,
		}
	}

	entrance, checkout := EntranceAndCheckout(layout)
	if entrance == nil || checkout == nil {
		return nil, errors.New("sort entries: unable to locate entrance or checkout")
	}

	zones := make([]*optishop.Zone, 0, len(list)+2)
	zones = append(zones, entrance)
	zoneSet := map[*optishop.Zone]bool{}
	for _, entry := range newList {
		if !zoneSet[entry.Info.Zone] {
			zoneSet[entry.Info.Zone] = true
			zones = append(zones, entry.Info.Zone)
		}
	}
	zones = append(zones, checkout)
	points := ZonesToPoints(layout, zones)
	distFunc := conn.DistanceFunc(points)
	solution := optishop.SolveTSP(len(points), distFunc)

	var result []*db.ListEntry
	for _, idx := range solution[1 : len(solution)-1] {
		zone := zones[idx]
		for _, x := range newList {
			if x.Info.Zone == zone {
				result = append(result, x)
			}
		}
	}

	return result, nil
}

// RoutePaths finds the optimal route and returns all of
// the path segments of it, as well as the sorted list of
// entries for convenience.
func RoutePaths(list []*db.ListEntry, store optishop.Store,
	conn *optishop.FloorConnector) ([]optishop.FloorPath, []*db.ListEntry, error) {
	entrance, checkout := EntranceAndCheckout(store.Layout())
	if entrance == nil || checkout == nil {
		return nil, nil, errors.New("route paths: unable to locate entrance or checkout")
	}

	sorted, err := SortEntries(list, store, conn)
	if err != nil {
		return nil, nil, errors.Wrap(err, "route paths")
	}

	zones := make([]*optishop.Zone, len(sorted)+2)
	zones[0] = entrance
	for i, e := range sorted {
		zones[i+1] = e.Info.Zone
	}
	zones[len(sorted)+1] = checkout
	points := ZonesToPoints(store.Layout(), zones)

	var res []optishop.FloorPath
	for i := 1; i < len(zones); i++ {
		path := conn.Connect(points[i-1], points[i])
		if path == nil {
			return nil, nil, errors.New("route paths: unable to connect two points")
		}
		res = append(res, path)
	}
	return res, sorted, nil
}

// EntranceAndCheckout finds the entrance and checkout
// zones for the layout, returning nil if they are not
// found.
func EntranceAndCheckout(l *optishop.Layout) (entrance *optishop.Zone, checkout *optishop.Zone) {
	for _, f := range l.Floors {
		for _, z := range f.Zones {
			if z.Checkout {
				checkout = z
			}
			if z.Entrance {
				entrance = z
			}
		}
	}
	return
}

func findExactZone(l *optishop.Layout, floor int, zone *optishop.Zone) *optishop.Zone {
	if floor < 0 || floor >= len(l.Floors) {
		return nil
	}
	for _, z := range l.Floors[floor].Zones {
		if z.Name == zone.Name && z.Location == zone.Location {
			return z
		}
	}
	// Fallback on name-only match.
	for _, z := range l.Floors[floor].Zones {
		if z.Name == zone.Name {
			return z
		}
	}
	return nil
}

// ZonesToPoints converts a slice of zones to a slice of
// floor-point pairs, suitable for passing to a connector.
func ZonesToPoints(l *optishop.Layout, zones []*optishop.Zone) []optishop.FloorPoint {
	points := make([]optishop.FloorPoint, len(zones))
	for i, z := range zones {
		if z == nil {
			continue
		}
		points[i] = optishop.FloorPoint{
			Floor: l.ZoneFloor(z),
			Point: z.Location,
		}
	}
	return points
}
