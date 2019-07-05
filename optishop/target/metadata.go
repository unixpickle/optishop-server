package target

import "encoding/json"

// ResponseMetadata is included in various responses and
// is stored in a slightly unusual way.
type RequestMetadata map[string]string

func (r RequestMetadata) MarshalJSON() ([]byte, error) {
	var fields []*metadataField
	for k, v := range r {
		fields = append(fields, &metadataField{Name: k, Value: v})
	}
	return json.Marshal(fields)
}

func (r RequestMetadata) UnmarshalJSON(data []byte) error {
	var fields []*metadataField
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for k := range r {
		delete(r, k)
	}
	for _, f := range fields {
		r[f.Name] = f.Value
	}
	return nil
}

type metadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
