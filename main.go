package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"path/filepath"

	svg "github.com/ajstarks/svgo/float"
	"github.com/unixpickle/optishop-server/optishop/visualize"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/optishop-server/optishop"
	"github.com/unixpickle/optishop-server/optishop/target"
)

func main() {
	var args Args
	args.Add()
	flag.Parse()

	store, err := target.NewStore(args.StoreID)
	essentials.Must(err)

	server := &Server{Store: store, AssetDir: args.AssetDir}
	http.HandleFunc("/", server.HandleRoot)
	http.HandleFunc("/script.js", server.HandleScript)
	http.HandleFunc("/style.css", server.HandleStyle)
	http.HandleFunc("/list", server.HandleList)
	http.HandleFunc("/search", server.HandleSearch)
	http.HandleFunc("/add", server.HandleAdd)
	http.HandleFunc("/delete", server.HandleDelete)
	http.HandleFunc("/route", server.HandleRoute)
	http.ListenAndServe(args.Addr, nil)
}

type Server struct {
	Store    optishop.Store
	AssetDir string

	List  []optishop.InventoryProduct
	Zones []*optishop.Zone
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.AssetDir, "index.html"))
}

func (s *Server) HandleScript(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.AssetDir, "script.js"))
}

func (s *Server) HandleStyle(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(s.AssetDir, "style.css"))
}

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	list := ConvertListItems(s.List)
	resultObj := map[string]interface{}{"list": list, "zones": s.Zones}
	json.NewEncoder(w).Encode(resultObj)
}

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	results, err := s.Store.Search(r.FormValue("query"))
	if err != nil {
		serveError(w, r, err)
		return
	}
	exactResults := make([][]byte, len(results))
	for i, res := range results {
		exactResults[i], err = s.Store.MarshalProduct(res)
		if err != nil {
			serveError(w, r, err)
			return
		}
	}
	resultObj := map[string]interface{}{
		"results":      ConvertListItems(results),
		"exactResults": exactResults,
	}
	json.NewEncoder(w).Encode(resultObj)
}

func (s *Server) HandleAdd(w http.ResponseWriter, r *http.Request) {
	rawItemData := r.FormValue("item")
	var itemData []byte
	if err := json.Unmarshal([]byte(rawItemData), &itemData); err != nil {
		serveError(w, r, err)
		return
	}
	product, err := s.Store.UnmarshalProduct(itemData)
	if err != nil {
		serveError(w, r, err)
		return
	}
	zone, err := s.Store.Locate(product)
	if err != nil {
		serveError(w, r, err)
		return
	}
	s.List = append(s.List, product)
	s.Zones = append(s.Zones, zone)

	s.HandleList(w, r)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	for i, prod := range s.List {
		if prod.Name() == name {
			essentials.OrderedDelete(&s.List, i)
			essentials.OrderedDelete(&s.Zones, i)
			s.HandleList(w, r)
			return
		}
	}
	serveError(w, r, errors.New("no list item found"))
}

func (s *Server) HandleRoute(w http.ResponseWriter, r *http.Request) {
	var points []optishop.FloorPoint

	layout := s.Store.Layout()
	entrance := layout.Zone("entrance")
	checkout := layout.Zone("checkout")
	if entrance == nil || checkout == nil {
		serveError(w, r, errors.New("missing either an entrance or a checkout"))
		return
	}
	points = append(points, optishop.FloorPoint{
		Point: entrance.Location,
		Floor: layout.ZoneFloor(entrance),
	})
	for i, item := range s.List {
		zone := s.Zones[i]
		if zone == nil {
			serveError(w, r, fmt.Errorf("could not locate: %s", item.Name()))
			return
		}
		points = append(points, optishop.FloorPoint{
			Point: zone.Location,
			Floor: layout.ZoneFloor(zone),
		})
	}
	points = append(points, optishop.FloorPoint{
		Point: checkout.Location,
		Floor: layout.ZoneFloor(checkout),
	})

	floorConn := optishop.NewFloorConnectorCached(layout)
	distFunc := floorConn.DistanceFunc(points)
	if distFunc == nil {
		serveError(w, r, errors.New("cannot reach all points from all other points"))
		return
	}
	solution := (optishop.FactorialTSPSolver{}).SolveTSP(len(points), distFunc)

	var sortedItems []optishop.InventoryProduct
	var sortedZones []*optishop.Zone
	for _, i := range solution[1 : len(solution)-1] {
		sortedItems = append(sortedItems, s.List[i-1])
		sortedZones = append(sortedZones, s.Zones[i-1])
	}

	resultObj := map[string]interface{}{
		"items": ConvertListItems(sortedItems),
		"zones": sortedZones,
		"image": s.renderSVG(points, solution, floorConn),
	}
	json.NewEncoder(w).Encode(resultObj)
}

func (s *Server) renderSVG(points []optishop.FloorPoint, solution []int,
	floorConn *optishop.FloorConnector) string {
	output := bytes.NewBuffer(nil)
	width, height, _ := visualize.MultiFloorGeometry(s.Store.Layout())

	canvas := svg.New(output)
	canvas.Start(width, height, fmt.Sprintf("viewbox=\"0 0 %.3f %.3f\"", width, height))

	visualize.DrawFloors(canvas, s.Store.Layout())
	for i := 0; i < len(solution)-1; i++ {
		path := floorConn.Connect(points[solution[i]], points[solution[i+1]])
		if path != nil {
			visualize.DrawFloorPath(canvas, s.Store.Layout(), path)
		}
	}

	canvas.End()

	return string(output.Bytes())
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	obj := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(obj)
}
