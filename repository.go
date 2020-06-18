package main

import "errors"

var ErrNoDatabaseMetaDataStored = errors.New("there is no database metadata stored in repository")

type Repository interface {
	GetDatabaseInfo() (databaseInfo, error)
	GetTables() (Tables, error)
	GetColumns() (ColumnsMetadata, error)
	UpdateAddTableDescription(tableID string, description string) error
	UpdateAddColumnDescription(columnID string, description string) error
	RemoveTable(tableID string) error
	RemoveColMetadata(colID string) error
	Setup
}

type Setup interface {
	AddDatabaseInfo(databaseInfo) error
	AddTable(table) error
	AddColMetaData(tableName string, col colMetadata) error
	RemoveEverything() error
	IsDatabaseMetaDataAdded(databaseName string) (bool, error)
}
