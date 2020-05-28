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
// If the primary key does not exist get() will return an error.
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
	Table       string
	TargetTable string
	Col         string
	DeleteRule  string
	UpdateRule  string
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
// If the foreign key does not exist get() will return an error.
func (fks ForeignKeys) get(colName string) (foreignKey, error) {
	for i := range fks {
		if fks[i].Col == colName {
			return fks[i], nil
		}
	}
	return foreignKey{}, errors.Errorf("foreign key with name %s does not exist", colName)
}

// colAndEnum holds information about a column and its enum type.
type colAndEnum struct {
	Table     string
	Col       string
	EnumName  string
	EnumValue string
}

// ColumnsAndEnums is a collection of columns with their corresponding enum types.
type ColumnsAndEnums []colAndEnum

// exists checks whether a column with the given colName exists in ColumnsAndEnums a or not.
func (ces ColumnsAndEnums) exists(colName string) bool {
	for i := range ces {
		if ces[i].Col == colName {
			return true
		}
	}
	return false
}

// get will get the column with the given colName and its enum type from ColumnsAndEnums.
// If the colName does not exist get() will return an error.
func (ces ColumnsAndEnums) get(colName string) (colAndEnum, error) {
	for i := range ces {
		if ces[i].Col == colName {
			return ces[i], nil
		}
	}
	return colAndEnum{}, errors.Errorf("column %s does not have any enum type.", colName)
}

// uniqueCol holds information about unique columns.
type uniqueCol struct {
	Table      string
	Col        string
	Definition string
}

// uniqueCols is a collection of columns with unique indexes.
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

// get will get the column with unique index with the given colName and tbName.
// If it does not exist get() will return an error.
func (ucs UniqueCols) get(colName string, tbName string) (uniqueCol, error) {
	for i := range ucs {
		if ucs[i].Col == colName && ucs[i].Table == tbName {
			return ucs[i], nil
		}
	}
	return uniqueCol{}, errors.Errorf("there is no column with the given name %s and in the given table %s"+
		" with a unique index.", colName, tbName)
}
