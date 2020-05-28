package main

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
}
