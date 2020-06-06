package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// mysqlVars represents the information needed to make a connection with a mysql database.
var mysqlVars = map[string]string{
	"user":     "",
	"password": "",
	"host":     "",
	"port":     "",
	"database": "",
}

// formatMysqlSource formats the mysqlVars map into a valid dataSourceName url using the given *Config
// so we can connect to a mysql database.
func formatMysqlSource(conf *Config) string {
	mysqlVars["user"] = conf.DatabaseUser
	mysqlVars["password"] = conf.DatabasePassword
	mysqlVars["host"] = conf.DatabaseHost
	mysqlVars["port"] = strconv.Itoa(conf.DatabasePort)
	mysqlVars["database"] = conf.DatabaseName
	format := mysqlDbSource
	for k, v := range mysqlVars {
		format = strings.Replace(format, "{"+k+"}", v, -1)
	}
	return format
}

// validateSqlDriver validates whether the given *dbDriver flag to manage the database is allowed or not.
func validateSqlDriver(conf *Config) error {
	allowed := false
	for i := range allowedDrivers {
		if conf.DatabaseDriver == allowedDrivers[i] {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("the given driver %s is not supported", conf.DatabaseDriver)
	}
	return nil
}

// parseNullableFromCol allows us to handle the *sql.ColumnType method Nullable().
// If Nullable() fails parseNullableFromCol will gracefully return false.
func parseNullableFromCol(col *sql.ColumnType) bool {
	if isNullable, ok := col.Nullable(); !ok {
		return false
	} else {
		return isNullable
	}
}

// parseLengthFromCol allows us to handle the *sql.ColumnType method Length().
// If Length() fails parseLengthFromCol will gracefully return 0 as the length of the column.
func parseLengthFromCol(col *sql.ColumnType) int64 {
	if length, ok := col.Length(); !ok {
		return 0
	} else {
		return length
	}
}

// getTableNames will get all table names of the database.
func getTableNames(conf *Config) ([]string, error) {
	tableNames := make([]string, 0)
	q := fmt.Sprintf(`
		SELECT TABLE_NAME as table_name
		FROM   information_schema.tables 
		WHERE  TABLE_TYPE = 'BASE TABLE'
			   AND TABLE_SCHEMA = '%s';
	`, conf.DatabaseSchema)

	rows, err := DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		tableNames = append(tableNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tableNames, nil
}

// getTableColumns will get all the columns of the given table as *sql.ColumnType.
func getTableColumns(tableName string, conf *Config) ([]*sql.ColumnType, error) {
	var q string

	if conf.DatabaseDriver == "postgres" {
		q = fmt.Sprintf(psqlQueryGetColumns, tableName)
	} else if conf.DatabaseDriver == "mysql" {
		q = fmt.Sprintf(mysqlQueryGetColumns, tableName)
	}

	rows, err := DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rows.ColumnTypes()
}

// getPrimaryKeys will get all columns of the database tables that are primary keys.
func getPrimaryKeys(conf *Config) (PrimaryKeys, error) {
	pks := make(PrimaryKeys, 0)
	var q string

	if conf.DatabaseDriver == "mysql" {
		q = fmt.Sprintf(mysqlQueryGetPks, conf.DatabaseSchema)
	} else if conf.DatabaseDriver == "postgres" {
		q = psqlQueryGetPKs
	}

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
func getForeignKeys(conf *Config) (ForeignKeys, error) {
	fks := make(ForeignKeys, 0)
	var q string

	if conf.DatabaseDriver == "mysql" {
		q = fmt.Sprintf(mysqlQueryGetFKs, conf.DatabaseSchema)
	} else if conf.DatabaseDriver == "postgres" {
		q = psqlQueryGetFKs
	}

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
func getColsAndEnums(conf *Config) (ColumnsAndEnums, error) {
	ces := make(ColumnsAndEnums, 0)
	var q string

	if conf.DatabaseDriver == "mysql" {
		q = fmt.Sprintf(mysqlQueryEnumTypesAndCols, conf.DatabaseSchema)
	} else if conf.DatabaseDriver == "postgres" {
		q = psqlQueryEnumTypesAndCols
	}

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

// getUniqueCols will get all columns of tables in the database that have a unique index.
func getUniqueCols(conf *Config) (UniqueCols, error) {
	ucs := make(UniqueCols, 0)
	var q string

	if conf.DatabaseDriver == "mysql" {
		q = fmt.Sprintf(mysqlQueryGetUniquesColumns, conf.DatabaseSchema)
	} else if conf.DatabaseDriver == "postgres" {
		q = fmt.Sprintf(psqlQueryGetUniquesColumns, conf.DatabaseSchema)
	}

	rows, err := DB.Query(q)
	if err != nil {
		return ucs, err
	}
	defer rows.Close()

	for rows.Next() {
		uc := uniqueCol{}
		if err := rows.Scan(&uc.Table, &uc.Col); err != nil {
			return ucs, err
		}
		ucs = append(ucs, uc)
	}

	if err := rows.Err(); err != nil {
		return ucs, err
	}

	return ucs, nil
}

// compareStoredDatabaseInfoWithConfig is a helper function that checks if the
// stored database info matches the configuration passed when running the application.
// If there is no match we might tell the client to use the -force_delete flag.
func compareStoredDatabaseInfoWithConf(dbInfo databaseInfo, conf *Config) (equal bool, message string) {
	differences := make([]string, 0)
	if dbInfo.User != conf.DatabaseUser {
		differences = append(differences, fmt.Sprintf("stored user %s != %s", dbInfo.User, conf.DatabaseUser))
	}
	if dbInfo.Password != conf.DatabasePassword {
		differences = append(differences, fmt.Sprintf("stored db password %s != %s", dbInfo.Password, conf.DatabasePassword))
	}
	if dbInfo.Name != conf.DatabaseName {
		differences = append(differences, fmt.Sprintf("stored db name %s != %s", dbInfo.Name, conf.DatabaseName))
	}
	if dbInfo.Driver != conf.DatabaseDriver {
		differences = append(differences, fmt.Sprintf("stored db driver %s != %s", dbInfo.Driver, conf.DatabaseDriver))
	}
	if dbInfo.Host != conf.DatabaseHost {
		differences = append(differences, fmt.Sprintf("stored db host %s != %s", dbInfo.Host, conf.DatabaseHost))
	}
	if dbInfo.Port != conf.DatabasePort {
		differences = append(differences, fmt.Sprintf("stored db port %d != %d", dbInfo.Port, conf.DatabasePort))
	}
	if dbInfo.Schema != conf.DatabaseSchema {
		differences = append(differences, fmt.Sprintf("stored db schema %s != %s", dbInfo.Schema, conf.DatabaseSchema))
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

var mysqlQueryGetPks = `
	SELECT sta.column_name, 
		   tab.table_name 
	FROM   information_schema.tables AS tab 
		   INNER JOIN information_schema.statistics AS sta 
				   ON sta.table_schema = tab.table_schema 
					  AND sta.table_name = tab.table_name 
					  AND sta.index_name = 'primary' 
	WHERE  tab.table_schema = '%s' 
	ORDER  BY tab.table_name;
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

var mysqlQueryGetFKs = `
	SELECT rf.table_name            AS origin_table_name, 
		   rf.referenced_table_name AS target_table_name, 
		   kcu.column_name, 
		   rf.delete_rule, 
		   rf.update_rule 
	FROM   information_schema.referential_constraints AS rf 
		   JOIN information_schema.table_constraints AS tc 
			 ON rf.constraint_name = tc.constraint_name 
		   JOIN information_schema.key_column_usage AS kcu 
			 ON kcu.constraint_name = tc.constraint_name 
	WHERE  rf.constraint_schema = '%s'; 
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

var mysqlQueryEnumTypesAndCols = `
	SELECT col.table_name  AS table_name, 
		   col.column_name AS column_name, 
		   col.data_type   AS enum_type, 
		   Regexp_replace(REPLACE(col.column_type, 'enum', ''), '[\)\(\']', '') 
	FROM   information_schema.columns AS col 
	WHERE  col.data_type = 'enum' 
		   AND col.table_schema = '%s'; 
`

var psqlQueryGetUniquesColumns = `
	SELECT DISTINCT 
           tbl.relname                     AS table_name, 
		   pga.attname                     AS column_name
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
`

var mysqlQueryGetUniquesColumns = `
	SELECT DISTINCT kcu.table_name  AS table_name, 
					kcu.column_name AS column_name
	FROM   information_schema.key_column_usage AS kcu 
	WHERE  kcu.table_schema = '%[1]s' 
		   AND kcu.constraint_name IN (SELECT tc.constraint_name 
									   FROM   information_schema.table_constraints 
											  AS tc 
									   WHERE  tc.constraint_type = 'UNIQUE' 
											  AND tc.table_schema = '%[1]s');
`

var mysqlQueryGetColumns = "SELECT * FROM `%s` LIMIT 0;"
var psqlQueryGetColumns = "SELECT * FROM %q LIMIT 0;"
