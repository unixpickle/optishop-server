package target

import (
	"github.com/pkg/errors"
	"github.com/unixpickle/optishop-server/optishop"
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
	// TODO: this.
	return &FloorDetails{}, nil
}
