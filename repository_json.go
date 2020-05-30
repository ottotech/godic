package main

import (
	"encoding/json"
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

func (s *jsonStorage) AddDatabaseInfo(info dbInfo) error {
	err := s.db.Write(db, "1", info)
	if err != nil {
		return err
	}
	return nil
}

func (s *jsonStorage) IsDatabaseMetaDataAdded(dbName string) (bool, error) {
	dbMeta := dbInfo{}
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

func (s *jsonStorage) AddTable(t table) error {
	t.ID = t.Name
	err := s.db.Write(collectionTable, t.Name, t)
	if err != nil {
		return errors.Errorf("got error while trying to add table %s in storage; %s", t.Name, err)
	}
	return nil
}

func (s *jsonStorage) AddColMetaData(tbName string, col colMetaData) error {
	ss, err := s.db.ReadAll(collectionColumn)
	if err != nil && !os.IsNotExist(err) {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tbName, err)
	}

	resource := tbName + "_" + col.Name + "_" + strconv.Itoa(len(ss)+1)
	col.ID = resource
	err = s.db.Write(collectionColumn, resource, col)
	if err != nil {
		return errors.Errorf("got error while trying to add column meta data of column %s in table %s; %s",
			col.Name, tbName, err)
	}

	return nil
}

func (s *jsonStorage) GetTables() (Tables, error) {
	tables := make(Tables, 0)
	list, err := s.db.ReadAll(collectionTable)
	if err != nil {
		return tables, err
	}
	for i := range list {
		var t table
		err := json.Unmarshal([]byte(list[i]), &t)
		if err != nil {
			return tables, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (s *jsonStorage) GetDatabaseInfo() (dbInfo, error) {
	info := dbInfo{}
	err := s.db.Read(db, "1", &info)
	if err != nil {
		if os.IsNotExist(err) {
			return info, ErrNoDatabaseMetaDataStored
		}
		return info, err
	}
	return info, nil
}

func (s *jsonStorage) RemoveEverything() error {
	err := os.RemoveAll("./data/*")
	if err != nil {
		return err
	}
	return nil
}
