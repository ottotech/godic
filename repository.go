package main

import "errors"

var ErrNoDatabaseMetaDataStored = errors.New("there is no database metadata stored in repository")

type Repository interface {
	AddDatabaseInfo(databaseInfo) error
	AddTable(table) error
	AddColMetaData(tbName string, col colMetaData) error
	UpdateAddTableDescription(tableID string, description string) error
	UpdateAddColumnDescription(columnID string, description string) error
	GetTables() (Tables, error)
	GetDatabaseInfo() (databaseInfo, error)
	IsDatabaseMetaDataAdded(dbName string) (bool, error)
	RemoveEverything() error
}
