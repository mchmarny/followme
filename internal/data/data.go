package data

import (
	"os/user"
	"path"

	"github.com/asdine/storm/v3"
	"github.com/pkg/errors"
)

const (
	dbFileName = ".followme.db"
)

// GetDB provides consistent way of obtaining DB
func GetDB() (*storm.DB, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, errors.Wrap(err, "error getting current user")
	}

	dbPath := path.Join(usr.HomeDir, dbFileName)
	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening DB: %s", dbPath)
	}

	return db, nil
}
