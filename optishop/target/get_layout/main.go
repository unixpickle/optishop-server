// Command get_layout gets the layout for a specific
// store and dumps it as JSON.
package main

import (
	"encoding/json"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop/target"
)

func main() {
	if len(os.Args) != 2 {
		essentials.Die("Usage: get_layout <store ID>")
	}
	storeID := os.Args[1]

	mapInfo, err := target.GetMapInfo(storeID)
	essentials.Must(err)

	layout := target.MapInfoToLayout(mapInfo)

	data, err := json.Marshal(layout)
	essentials.Must(err)

	os.Stdout.Write(data)
}
