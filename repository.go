package main

type Repository interface {
	AddDatabaseInfo(dbInfo) error
	AddTable(table) error
	AddColMetaData(tbName string, col colMetaData) error
	IsDatabaseMetaDataAdded(dbName string) (bool, error)
	GetTables() (Tables, error)
}
