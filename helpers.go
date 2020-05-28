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

// getPrimaryKeys will get all columns of the DB tables that are primary keys.
func getPrimaryKeys() (PrimaryKeys, error) {
	pks := make(PrimaryKeys, 0)
	q := fmt.Sprintf(psqlQueryGetPKs)
	rows, err := DB.Query(q)
	if err != nil {
		return pks, err
	}
	defer rows.Close()

	for rows.Next() {
		pk := primaryKey{}
		if err := rows.Scan(&pk.Col, &pk.Table); err != nil {
			return pks, err
		}
		pks = append(pks, pk)
	}

	if err := rows.Err(); err != nil {
		return pks, err
	}

	return pks, nil
}

// getForeignKeys will get all columns of the DB tables that are foreign keys.
func getForeignKeys() (ForeignKeys, error) {
	fks := make(ForeignKeys, 0)
	q := psqlQueryGetFKs
	rows, err := DB.Query(q)
	if err != nil {
		return fks, err
	}
	defer rows.Close()

	for rows.Next() {
		fk := foreignKey{}
		if err := rows.Scan(&fk.Table, &fk.TargetTable, &fk.Col, &fk.DeleteRule, &fk.UpdateRule); err != nil {
			return fks, err
		}
		fks = append(fks, fk)
	}

	if err := rows.Err(); err != nil {
		return fks, err
	}

	return fks, nil
}

// getColsAndEnums will get all columns and their corresponding enum types from DB.
func getColsAndEnums() (ColumnsAndEnums, error) {
	ces := make(ColumnsAndEnums, 0)
	q := psqlQueryEnumTypesAndCols
	rows, err := DB.Query(q)
	if err != nil {
		return ces, err
	}
	defer rows.Close()

	for rows.Next() {
		ce := colAndEnum{}
		if err := rows.Scan(&ce.Table, &ce.Col, &ce.EnumName, &ce.EnumValue); err != nil {
			return ces, err
		}
		ces = append(ces, ce)
	}

	if err := rows.Err(); err != nil {
		return ces, err
	}

	return ces, nil
}

var psqlQueryGetPKs = `
	SELECT cu.column_name, 
		   cu.table_name 
	FROM   information_schema.key_column_usage AS cu 
		   JOIN information_schema.table_constraints AS tc 
			 ON tc.constraint_name = cu.constraint_name 
	WHERE  tc.constraint_type = 'PRIMARY KEY'; 
`

var psqlQueryGetFKs = `
	SELECT cu.table_name  AS origin_table_name, 
		   icu.table_name AS target_table_name, 
		   cu.column_name, 
		   rc.delete_rule, 
		   rc.update_rule 
	FROM   information_schema.key_column_usage AS cu 
		   JOIN information_schema.table_constraints AS tc 
			 ON tc.constraint_name = cu.constraint_name 
		   JOIN information_schema.referential_constraints AS rc 
			 ON tc.constraint_name = rc.constraint_name 
		   JOIN information_schema.constraint_column_usage AS icu 
			 ON icu.constraint_name = rc.constraint_name 
	WHERE  tc.constraint_type = 'FOREIGN KEY'; 
`

var psqlQueryEnumTypesAndCols = `
	SELECT isc.table_name, 
		   isc.column_name, 
		   t.typname                     AS enum_name, 
		   String_agg(e.enumlabel, ', ') AS enum_value 
	FROM   pg_type AS t 
		   JOIN pg_enum e 
			 ON t.oid = e.enumtypid 
		   JOIN pg_catalog.pg_namespace n 
			 ON n.oid = t.typnamespace 
		   JOIN information_schema.columns AS isc 
			 ON isc.udt_name = t.typname 
	GROUP  BY enum_name, 
			  isc.column_name, 
			  isc.table_name; 
`
