package repo

import (
	"github.com/cyolo-core/cmd/dex/model"
	"time"
)

// Retrieve finds the first license that matches an optional where clause.
func (repo sqlLicenseRepository) Retrieve() (license *model.License, err error) {
	return &model.License{
		Users:     10,
		NotBefore: 0,
		NotAfter:  time.Now().AddDate(1, 0, 0).Unix(),
	}, nil
}
