package model

import (
	"encoding/json"
	"strconv"
)

// Identity is a hack to overcome limitations in certain react.js components
// that require object-ids to be strings.
type Identity int

// MarshalJSON implements json.Marshaller to play nice with react.js.
func (id Identity) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(id)))
}

// Unmarshal implements json.Unmarshaler in order to parse out a numerical
// Identity value from its string representation.
func (id *Identity) UnmarshalJSON(buf []byte) (err error) {
	var (
		intId  int
		strBuf string
	)

	if err = json.Unmarshal(buf, &strBuf); err != nil {
		return
	}
	if intId, err = strconv.Atoi(strBuf); err != nil {
		return
	}

	*id = Identity(intId)
	return
}
