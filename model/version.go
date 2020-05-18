package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LatestVersion is the latest version of the code-base.
var LatestVersion = Version{1, 0, 1}

// Version represents a product version.
type Version struct{ Major, Minor, Patch int }

// String returns the string representation of version v.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// MarshalJSON returns the JSON representation of version.
func (v Version) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", v.String())), nil
}

// Unmarshal attempts to parse a version struct from data.
func (v *Version) UnmarshalJSON(data []byte) (err error) {
	strVersion := ""
	if err = json.Unmarshal(data, &strVersion); err != nil {
		return
	}
	if parts := strings.Split(strVersion, "."); len(parts) == 3 {
		if v.Major, err = strconv.Atoi(parts[0]); err != nil {
			return
		}
		if v.Minor, err = strconv.Atoi(parts[1]); err != nil {
			return
		}
		if v.Patch, err = strconv.Atoi(parts[2]); err != nil {
			return
		}
		return
	}
	return fmt.Errorf("malformed version struct: %q", data)
}
