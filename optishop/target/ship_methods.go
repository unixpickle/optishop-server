package target

// ShipMethodsResult is a legacy data structure describing
// where and how an item can be purchased or shipped.
//
// This is only retained because it may be stored in old
// user accounts, but it will never be returned by an API
// that fetches information from Target's website.
type ShipMethodsResult struct {
	ProductID             string `json:"product_id"`
	AvailabilityStatus    string `json:"availability_status"`
	PreferredStoreOptions []struct {
		PreferredStoreOptionID string `json:"preferred_store_option_id"`
		LocationID             string `json:"location_id"`
	} `json:"preferred_store_options"`
}

// InStore checks if the corresponding item is available
// in stores.
func (s *ShipMethodsResult) InStore() bool {
	for _, opt := range s.PreferredStoreOptions {
		if opt.PreferredStoreOptionID == "IN_STORE" {
			return true
		}
	}
	return false
}
