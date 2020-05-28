package main

import "github.com/pkg/errors"

// dbInfo holds information about the DB.
type dbInfo struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Driver   string `json:"driver"`
}

// table represents a table in DB.
type table struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tables is a collection of tables.
type Tables []table

// Count counts the number of tables the DB has.
func (t Tables) Count() int {
	return len(t)
}

// colMetaData holds meta data about a specific column in a table from DB.
type colMetaData struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	DBType       string `json:"db_type"`
	Nullable     bool   `json:"nullable"`
	GoType       string `json:"go_type"`
	Length       int64  `json:"length"`
	TBName       string `json:"table_name"`
	Description  string `json:"description"`
	IsPrimaryKey bool   `json:"is_primary_key"`
	IsForeignKey bool   `json:"is_foreign_key"`
	DeleteRule   string `json:"delete_rule"`
	UpdateRule   string `json:"update_rule"`
}

// primaryKey holds information about a primary key.
type primaryKey struct {
	Table string
	Col   string
}

// PrimaryKeys is a collection of primary keys.
type PrimaryKeys []primaryKey

// exists checks whether a primary key with the given colName exists in PrimaryKeys or not.
func (pks PrimaryKeys) exists(colName string) bool {
	for i := range pks {
		if pks[i].Col == colName {
			return true
		}
	}
	return false
}

// get will get the primary key with the given colName from PrimaryKeys.
// if the primary key does not exist get() will return an error.
func (pks PrimaryKeys) get(colName string) (primaryKey, error) {
	for i := range pks {
		if pks[i].Col == colName {
			return pks[i], nil
		}
	}
	return primaryKey{}, errors.Errorf("primary key with name %s does not exist", colName)
}


// foreignKey holds information about a foreign key.
type foreignKey struct {
	Table      string
	Col        string
	DeleteRule string
	UpdateRule string
}

// ForeignKeys is a collection of foreign keys.
type ForeignKeys []foreignKey

// exists checks whether a foreign key with the given colName exists in ForeignKeys or not.
func (fks ForeignKeys) exists(colName string) bool {
	for i := range fks {
		if fks[i].Col == colName {
			return true
		}
	}
	return false
}

// get will get the foreign key with the given colName from  ForeignKeys.
// if the foreign key does not exist get() will return an error.
func (fks ForeignKeys) get(colName string) (foreignKey, error) {
	for i := range fks {
		if fks[i].Col == colName {
			return fks[i], nil
		}
	}
	return foreignKey{}, errors.Errorf("foreign key with name %s does not exist", colName)
}
