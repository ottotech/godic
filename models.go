package main

import "github.com/pkg/errors"

// databaseInfo holds general information about the database.
type databaseInfo struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Driver   string `json:"driver"`
	Schema   string `json:"schema"`
}

// table represents a table in database.
type table struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Tables is a collection of tables.
type Tables []table

// Count counts the number of tables of the database.
func (t Tables) Count() int {
	return len(t)
}

// colMetaData holds meta data about a specific column in a table from database.
type colMetaData struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	DBType                string   `json:"db_type"`
	Nullable              bool     `json:"nullable"`
	GoType                string   `json:"go_type"`
	Length                int64    `json:"length"`
	TBName                string   `json:"table_name"`
	Description           string   `json:"description"`
	IsPrimaryKey          bool     `json:"is_primary_key"`
	IsForeignKey          bool     `json:"is_foreign_key"`
	TargetTableFK         string   `json:"target_table_fk"`
	DeleteRule            string   `json:"delete_rule"`
	UpdateRule            string   `json:"update_rule"`
	HasENUM               bool     `json:"has_enum"`
	ENUMName              string   `json:"enum_name"`
	ENUMValues            []string `json:"enum_values"`
	IsUnique              bool     `json:"is_unique"`
	UniqueIndexDefinition string   `json:"unique_index_definition"`
}

// primaryKey holds information about a primary key.
type primaryKey struct {
	Table string
	Col   string
}

// PrimaryKeys is a collection of primary keys.
type PrimaryKeys []primaryKey

// exists checks whether a primary key column with the given colName exists
// in the given tableName or not.
func (pks PrimaryKeys) exists(colName string, tableName string) bool {
	for i := range pks {
		if pks[i].Col == colName && pks[i].Table == tableName {
			return true
		}
	}
	return false
}

// get will get the primary key with the given colName in the given tableName.
// If the primary key does not exist get() will return an error.
func (pks PrimaryKeys) get(colName string, tableName string) (primaryKey, error) {
	for i := range pks {
		if pks[i].Col == colName && pks[i].Table == tableName {
			return pks[i], nil
		}
	}
	return primaryKey{}, errors.Errorf("primary key with name %s does not exist", colName)
}

// foreignKey holds information about a foreign key.
type foreignKey struct {
	Table       string
	TargetTable string
	Col         string
	DeleteRule  string
	UpdateRule  string
}

// ForeignKeys is a collection of foreign keys.
type ForeignKeys []foreignKey

// exists checks whether a foreign key with the given colName in the given tableName exists or not in ForeignKeys.
func (fks ForeignKeys) exists(colName string, tableName string) bool {
	for i := range fks {
		if fks[i].Col == colName && fks[i].Table == tableName {
			return true
		}
	}
	return false
}

// get will get the foreign key with the given colName in the given tableName.
// If the foreign key does not exist get() will return an error.
func (fks ForeignKeys) get(colName string, tableName string) (foreignKey, error) {
	for i := range fks {
		if fks[i].Col == colName && fks[i].Table == tableName {
			return fks[i], nil
		}
	}
	return foreignKey{}, errors.Errorf("foreign key with name %s does not exist", colName)
}

// colAndEnum holds information about a column and its enum type.
type colAndEnum struct {
	Table      string
	Col        string
	EnumName   string
	EnumValues string
}

// ColumnsAndEnums is a collection of columns with their corresponding enum types.
type ColumnsAndEnums []colAndEnum

// exists checks whether a column with the given colName in the given tableNae exists in ColumnsAndEnums or not.
func (ces ColumnsAndEnums) exists(colName string, tableName string) bool {
	for i := range ces {
		if ces[i].Col == colName && ces[i].Table == tableName {
			return true
		}
	}
	return false
}

// get will get the column with the given colName and its enum type from the given tableName.
// If the colName does not exist get() will return an error.
func (ces ColumnsAndEnums) get(colName string, tableName string) (colAndEnum, error) {
	for i := range ces {
		if ces[i].Col == colName && ces[i].Table == tableName {
			return ces[i], nil
		}
	}
	return colAndEnum{}, errors.Errorf("there is no column %s in table %s with an enum type.", colName, tableName)
}

// uniqueCol holds columns with an unique index.
type uniqueCol struct {
	Table string
	Col   string
}

// UniqueCols is a collection of columns with unique indexes.
type UniqueCols []uniqueCol

// exists checks whether a column with the given colName in the given tbName and with an unique index exists or not.
func (ucs UniqueCols) exists(colName string, tbName string) bool {
	for i := range ucs {
		if ucs[i].Col == colName && ucs[i].Table == tbName {
			return true
		}
	}
	return false
}

// get will get the column with unique index with the given colName from the given tableName.
// If it does not exist get() will return an error.
func (ucs UniqueCols) get(colName string, tableName string) (uniqueCol, error) {
	for i := range ucs {
		if ucs[i].Col == colName && ucs[i].Table == tableName {
			return ucs[i], nil
		}
	}
	return uniqueCol{}, errors.Errorf("there is no column with an unique index with the given name %s and in the "+
		"given table %s.", colName, tableName)
}
