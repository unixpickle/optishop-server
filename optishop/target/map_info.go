package target

import (
	"encoding/json"
	"math"

	"github.com/pkg/errors"
	"github.com/unixpickle/approb"
	"github.com/unixpickle/essentials"
)

// MapInfo contains generic information about the layout
// of a store and the parts of that store.
// It does not contain information about Aisles, however.
type MapInfo struct {
	StoreID           string      `json:"storeId"`
	StoreName         string      `json:"storeName"`
	StoreUnderRemodel bool        `json:"storeUnderRemodel"`
	FloorMapDetails   []*FloorMap `json:"floorMapDetails"`
}

// FloorMap holds the general layout of a single floor.
type FloorMap struct {
	FloorID    string       `json:"floorId"`
	MapMarkers []*MapMarker `json:"mapMarkers"`
}

// Portals gets the MapMarkers that are portals.
func (f *FloorMap) Portals() []*MapMarker {
	var res []*MapMarker
	for _, m := range f.MapMarkers {
		if m.Portal() {
			res = append(res, m)
		}
	}
	return res
}

// MapMarker is an indicator of some specific location on
// a specific floor.
type MapMarker struct {
	Name     string `json:"name"`
	IconName string `json:"iconName"`
	Type     string `json:"type"`
	Location struct {
		CenterX float64 `json:"centerX"`
		CenterY float64 `json:"centerY"`
		FloorID string  `json:"floorId"`
	} `json:"location"`
}

// Portal checks if the map marker is a way to get between
// floors of the store.
func (m *MapMarker) Portal() bool {
	return m.Name == "elevators" || m.Name == "escalators"
}

// GetMapInfo looks up general map information for a given
// store identifier.
func GetMapInfo(storeID string) (*MapInfo, error) {
	data, err := GetRequest("https://prod.tgtneptune.com/v2/stores/" + storeID + "/maps")
	if err != nil {
		return nil, errors.Wrap(err, "get map info")
	}

	var errObj struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &errObj); err == nil && errObj.Error != "" {
		return nil, errors.New("get map info: " + errObj.Error)
	}

	var info MapInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, errors.Wrap(err, "get map info")
	}
	return &info, nil
}

// MatchFloorPortals matches elevators and escalators
// between two consecutive floors, finding a sequence
// of pairs of {first_portal, second_portal} where the
// first floor's map marker is matched to the second's.
func MatchFloorPortals(first, second *FloorMap) [][2]*MapMarker {
	markers1 := first.Portals()
	markers2 := second.Portals()
	maxPairs := essentials.MinInt(len(markers1), len(markers2))
	for numPairs := maxPairs; numPairs > 0; numPairs-- {
		var bestMatch [][2]*MapMarker
		bestLoss := math.Inf(1)
		for _, perm1 := range markerPerms(markers1) {
			for _, perm2 := range markerPerms(markers2) {
				loss := matchLoss(perm1[:numPairs], perm2[:numPairs])
				if loss < bestLoss {
					bestLoss = loss
					bestMatch = make([][2]*MapMarker, numPairs)
					for i := range bestMatch {
						bestMatch[i] = [2]*MapMarker{perm1[i], perm2[i]}
					}
				}
			}
		}
		if bestMatch != nil {
			return bestMatch
		}
	}
	return nil
}

func markerPerms(m []*MapMarker) [][]*MapMarker {
	var res [][]*MapMarker
	for perm := range approb.Perms(len(m)) {
		markers := make([]*MapMarker, len(m))
		for i, j := range perm {
			markers[i] = m[j]
		}
		res = append(res, markers)
	}
	return res
}

// matchLoss computes a value which is lower for better
// matches.
func matchLoss(m1, m2 []*MapMarker) float64 {
	for i, marker := range m1 {
		if marker.Name != m2[i].Name {
			// Cannot match different types of portals.
			return math.Inf(1)
		}
	}

	// Find the optimal x,y shift to reduce MSE.
	var meanX, meanY float64
	for i, marker := range m1 {
		meanX += m2[i].Location.CenterX - marker.Location.CenterX
		meanY += m2[i].Location.CenterY - marker.Location.CenterY
	}
	meanX /= float64(len(m1))
	meanY /= float64(len(m2))

	var squareDist float64
	for i, marker := range m1 {
		xDist := math.Pow(marker.Location.CenterX+meanX-m2[i].Location.CenterX, 2)
		yDist := math.Pow(marker.Location.CenterY+meanY-m2[i].Location.CenterY, 2)
		squareDist += xDist + yDist
	}
	return squareDist
}
