package main

import (
	"database/sql"
	"fmt"
	"strings"
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
		if err := rows.Scan(&ce.Table, &ce.Col, &ce.EnumName, &ce.EnumValues); err != nil {
			return ces, err
		}
		ces = append(ces, ce)
	}

	if err := rows.Err(); err != nil {
		return ces, err
	}

	return ces, nil
}

func getUniqueCols() (UniqueCols, error) {
	ucs := make(UniqueCols, 0)
	q := psqlQueryGetUniquesColumns
	rows, err := DB.Query(q)
	if err != nil {
		return ucs, err
	}
	defer rows.Close()

	for rows.Next() {
		uc := uniqueCol{}
		if err := rows.Scan(&uc.Table, &uc.Col, &uc.UniqueDefinition); err != nil {
			return ucs, err
		}
		ucs = append(ucs, uc)
	}

	if err := rows.Err(); err != nil {
		return ucs, err
	}

	return ucs, nil
}

// compareStoredDatabaseInfoWithFlags is a helper function that checks if the
// stored database info matches the flags passed when running the application.
// If there is no match we might tell the client to use the -update or -force_delete
// flags.
func compareStoredDatabaseInfoWithFlags(info dbInfo) (equal bool, message string) {
	differences := make([]string, 0)
	if info.User != *dbUser {
		differences = append(differences, fmt.Sprintf("stored user %s != %s", info.User, *dbUser))
	}
	if info.Password != *dbPassword {
		differences = append(differences, fmt.Sprintf("stored db password %s != %s", info.Password, *dbPassword))
	}
	if info.Name != *dbName {
		differences = append(differences, fmt.Sprintf("stored db name %s != %s", info.Name, *dbName))
	}
	if info.Driver != *dbDriver {
		differences = append(differences, fmt.Sprintf("stored db driver %s != %s", info.Driver, *dbDriver))
	}
	if info.Host != *dbHost {
		differences = append(differences, fmt.Sprintf("stored db host %s != %s", info.Host, *dbHost))
	}
	if info.Port != *dbPort {
		differences = append(differences, fmt.Sprintf("stored db port %d != %d", info.Port, *dbPort))
	}
	if len(differences) > 0 {
		message = strings.Join(differences, ".\n")
	} else if len(differences) == 0 {
		equal = true
	}
	return
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

var psqlQueryGetUniquesColumns = fmt.Sprintf(`
	SELECT tbl.relname                     AS table_name, 
		   pga.attname                     AS column_name, 
		   Pg_get_indexdef(pgi.indexrelid) AS definition 
	FROM   pg_index AS pgi 
		   JOIN pg_class AS pgc 
			 ON pgc.oid = pgi.indexrelid 
		   JOIN pg_namespace AS pgn 
			 ON pgn.oid = pgc.relnamespace 
		   JOIN pg_class AS tbl 
			 ON tbl.oid = pgi.indrelid 
		   JOIN pg_attribute AS pga 
			 ON pga.attrelid = pgc.oid 
	WHERE  pgi.indisunique = true 
		   AND pgn.nspname = '%s'; 
`, *dbSchema)
