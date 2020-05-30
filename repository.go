package main

import "errors"

var ErrNoDatabaseMetaDataStored = errors.New("there is no database meta data stored in repository")

type Repository interface {
	AddDatabaseInfo(dbInfo) error
	AddTable(table) error
	AddColMetaData(tbName string, col colMetaData) error
	GetTables() (Tables, error)
	GetDatabaseInfo() (dbInfo, error)
	IsDatabaseMetaDataAdded(dbName string) (bool, error)
	RemoveEverything() error
}
