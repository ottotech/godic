package main

import (
	"database/sql"
	"fmt"
)

func parseNullableFromCol(col *sql.ColumnType) bool {
	if isNullable, ok := col.Nullable(); !ok {
		return false
	} else {
		return isNullable
	}
}

func parseLengthFromCol(col *sql.ColumnType) int64 {
	if length, ok := col.Length(); !ok {
		return 0
	} else {
		return length
	}
}

// getPrimaryKeys will get all columns of the DB tables that are
// primary keys. The returned map will have keys representing the
// the column names and values representing the table names.
func getPrimaryKeys() (map[string]string, error) {
	m := make(map[string]string)
	q := fmt.Sprintf(queryKeyColUsage, pk)
	rows, err := DB.Query(q)
	if err != nil {
		return m, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			col   string
			table string
		)
		if err := rows.Scan(&col, &table); err != nil {
			return m, err
		}
		m[col] = table
	}

	if err := rows.Err(); err != nil {
		return m, err
	}

	return m, nil
}

// getForeignKeys will get all columns of the DB tables that are
// foreign keys. The returned map will have keys representing the
// the column names and values representing the table names.
func getForeignKeys() (map[string]string, error) {
	m := make(map[string]string)
	q := fmt.Sprintf(queryKeyColUsage, fk)
	rows, err := DB.Query(q)
	if err != nil {
		return m, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			col, table string
		)
		if err := rows.Scan(&col, &table); err != nil {
			return m, err
		}
		m[col] = table
	}

	if err := rows.Err(); err != nil {
		return m, err
	}

	return m, nil
}

var queryKeyColUsage = `
	SELECT cu.column_name, 
		   cu.table_name 
	FROM   information_schema.key_column_usage AS cu 
		   JOIN information_schema.table_constraints AS tc 
			 ON tc.constraint_name = cu.constraint_name 
	WHERE  tc.constraint_type = '%s'; 
`
