package main

type dbInfo struct {
	Name     string `json:"name"`
	User     string `json:"user"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Driver   string `json:"driver"`
}

type colMetaData struct {
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

// table represents a table in DB.
type table struct {
	Name        string `json:""`
	Description string `json:"description"`
}

// Tables is a collection of tables.
type Tables []table

// Count counts the number of tables the DB has.
func (t Tables) Count() int {
	return len(t)
}
