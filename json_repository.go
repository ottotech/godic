package main

import (
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/pkg/errors"
	"os"
	"strconv"
)

const (
	// dir defines the name of the directory where the files are stored.
	dir = "./data/"

	// collectionTable identifier for the JSON collection of tables.
	collectionTable = "tables"

	// collectionColumn identifier for the JSON collection of columns.
	collectionColumn = "columns"

	// db identifier for the database info.
	db = "db"
)

//jsonStorage stores the data in json files.
type jsonStorage struct {
	db *scribble.Driver
}

// NewJsonStorage returns a json storage.
func NewJsonStorage() (*jsonStorage, error) {
	var err error
	s := new(jsonStorage)
	s.db, err = scribble.New(dir, nil)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *jsonStorage) AddDB(data dbMetaData) error {
	err := s.db.Write(db, "1", data)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) IsDBAdded(dbName string) (bool, error) {
	dbMeta := dbMetaData{}
	err := s.db.Read(db, "1", &dbMeta)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if dbMeta.Name != dbName {
		return false, nil
	}
	return true, nil
}

func (s *jsonStorage) AddTable(name string) error {
	err := s.db.Write(collectionTable, name, name)
	if err != nil {
		return errors.Errorf("got error while trying to add table %s in storage; %s", name, err)
	}
	return nil
}

func (s *jsonStorage) AddDBTableColMetaData(tbName string, col colMetaData) error {
	ss, err := s.db.ReadAll(collectionColumn)
	if err != nil && !os.IsNotExist(err) {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tbName, err)
	}

	resource := tbName + "_" + col.Name + "_" + strconv.Itoa(len(ss)+1)
	err = s.db.Write(collectionColumn, resource, col)
	if err != nil {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tbName, err)
	}

	return nil
}
