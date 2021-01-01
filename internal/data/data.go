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

// GetDefaultDBFilePath uses current user to build default file path
func GetDefaultDBFilePath() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}

	return path.Join(usr.HomeDir, dbFileName)
}

// GetDB provides consistent way of obtaining DB
func GetDB(dbPath string) (*storm.DB, error) {
	if dbPath == "" {
		dbPath = GetDefaultDBFilePath()
	}

	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, errors.Wrapf(err, "error opening DB: %s", dbPath)
	}

	return db, nil
}
