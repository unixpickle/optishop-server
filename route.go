package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/db"
)

// SortEntries finds the optimal route and returns the
// list entries sorted by this route.
func SortEntries(list []*db.ListEntry, store optishop.Store) ([]*db.ListEntry, error) {
	layout := store.Layout()
	newList := make([]*db.ListEntry, len(list))
	for i, entry := range list {
		info := *entry.Info
		info.Zone = findExactZone(layout, info.Floor, info.Zone)
		if info.Zone == nil {
			return nil, fmt.Errorf("sort entries: invalid zone for list entry %d", i)
		}
		newList[i] = &db.ListEntry{
			ID:   entry.ID,
			Info: &info,
		}
	}

	entrance, checkout := EntranceAndCheckout(layout)
	if entrance == nil || checkout == nil {
		return nil, errors.New("sort entries: unable to locate entrance or exit")
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
	points := zonesToPoints(layout, zones)
	floorConn := optishop.NewFloorConnectorCached(layout)
	distFunc := floorConn.DistanceFunc(points)
	solution := (optishop.FactorialTSPSolver{}).SolveTSP(len(points), distFunc)
	// TODO: sort according to TSP.
	return nil, errors.New("sort entries: not yet implemented")
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
	return nil
}

func zonesToPoints(l *optishop.Layout, zones []*optishop.Zone) []optishop.FloorPoint {
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
