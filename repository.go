package main

type Repository interface {
	AddDB(dbInfo) error
	AddTable(table) error
	AddColMetaData(tbName string, col colMetaData) error
	IsDBAdded(dbName string) (bool, error)
	GetTables() (Tables, error)
}
