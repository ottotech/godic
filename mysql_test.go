package main

import (
	"database/sql"
	"reflect"
	"testing"
)

func Test_getTableNames_helper_func_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	tables, err := getTableNames(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error from getTableNames")
	}

	expectedTables := []string{"order", "product", "order_line"}

	for _, e := range expectedTables {
		exists := false
		for _, t := range tables {
			if e == t {
				exists = true
				break
			}
		}
		if !exists {
			t.Errorf("expected table name %s.", e)
		}
	}
}

func Test_getTableColumns_helpers_func_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	tables, err := getTableNames(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error from getTableNames")
	}

	// A map of tables (keys) and its columns (values as list of columns)
	expectations := map[string][]string{
		"order":      {"id"},
		"product":    {"id", "name", "counting_option"},
		"order_line": {"id", "order_id"},
	}

	for i := range tables {
		cols, err := getTableColumns(tables[i], conf)
		if err != nil {
			t.Fatalf("we shouldn't get an error from getTableColumns")
		}
		for _, e := range expectations[tables[i]] {
			exists := false
			for _, col := range cols {
				if col.Name() == e {
					exists = true
					break
				}
			}
			if !exists {
				t.Errorf("expected column %s in table %s", e, tables[i])
			}
		}
	}
}

func Test_getPrimaryKeys_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	pks, err := getPrimaryKeys(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error from getPrimaryKeys; got %s", err)
	}

	if len(pks) != 3 {
		t.Fatalf("expected 3 primary keys got %d", len(pks))
	}

	expectedPrimaryKeys := []primaryKey{
		{
			Table: "order",
			Col:   "id",
		},
		{
			Table: "product",
			Col:   "id",
		},
		{
			Table: "order_line",
			Col:   "id",
		},
	}

	for _, e := range expectedPrimaryKeys {
		exists := false
		for _, pk := range pks {
			if pk.Table == e.Table && pk.Col == e.Col {
				exists = true
				break
			}
		}
		if !exists {
			t.Errorf("expected primary key %s in table %s", e.Col, e.Table)
		}
	}
}

func Test_getForeignKeys_helper_func_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	fks, err := getForeignKeys(conf)
	if err != nil {
		t.Errorf("we shouldn't get an error from getForeignKeys; got %s", err)
	}

	if len(fks) != 2 {
		t.Errorf("we expected 2 foreign keys got %d", len(fks))
	}

	expectedForeignKeys := []foreignKey{
		{
			Table:       "order_line",
			TargetTable: "order",
			Col:         "order_id",
			DeleteRule:  "RESTRICT",
			UpdateRule:  "NO ACTION",
		}, {
			Table:       "order_line",
			TargetTable: "product",
			Col:         "product_id",
			DeleteRule:  "RESTRICT",
			UpdateRule:  "NO ACTION",
		},
	}

	for _, e := range expectedForeignKeys {
		exists := false
		for _, fk := range fks {
			if fk.Table == e.Table && fk.Col == e.Col {
				exists = true
				if !reflect.DeepEqual(fk, e) {
					t.Errorf("expected fk (%+v); got %+v", e, fk)
				}
			}
		}
		if !exists {
			t.Errorf("expected fk %s in table %s", e.Col, e.Table)
		}
	}
}

func Test_getColsAndEnums_helper_func_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	enums, err := getColsAndEnums(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error when calling getColsAndEnums; got %s", err)
	}

	if len(enums) != 1 {
		t.Errorf("expected to have 1 enum only; got %d", len(enums))
	}

	expectedEnum := colAndEnum{
		Table:      "product",
		Col:        "counting_option",
		EnumName:   "enum",
		EnumValues: "unit,decimal",
	}

	if !reflect.DeepEqual(expectedEnum, enums[0]) {
		t.Errorf("expected enum %+v; got %+v", expectedEnum, enums[0])
	}
}

func Test_getUniqueCols_helper_func_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()

	uniqueColumns, err := getUniqueCols(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error when calling getUniqueCols; got %s", err)
	}

	if len(uniqueColumns) != 4 {
		t.Errorf("we expected 4 unique columns got %d", len(uniqueColumns))
	}

	expectedUniqueColumns := []uniqueCol{
		{
			Table: "order_line",
			Col:   "id",
		}, {
			Table: "order",
			Col:   "id",
		}, {
			Table: "product",
			Col:   "id",
		}, {
			Table: "product",
			Col:   "name",
		},
	}

	for _, e := range expectedUniqueColumns {
		exists := false
		for _, unique := range uniqueColumns {
			if unique.Table == e.Table && unique.Col == e.Col {
				exists = true
				break
			}
		}
		if !exists {
			t.Errorf("expected to have a unique column %s in table %s", e.Col, e.Table)
		}
	}
}

func Test_databaseMetaDataSetup_AND_some_repository_methods_for_mysql_db(t *testing.T) {
	originalDB := DB
	DB = mysqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createMysqlConf()
	conf.ForceDelete = true // for testing we force delete of the data in the database.

	storage, err := NewJsonStorage()
	if err != nil {
		t.Fatalf("we shouldn't get an error from NewJsonStorage; got %s", err)
	}

	// Test setup.

	err = setupInitialMetadata(storage, conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error from databaseMetaDataSetup; got %s", err)
	}

	// Test some repository methods.

	tables, err := storage.GetTables()
	if err != nil {
		t.Fatalf("we shouldn't get an error from GetTables; got %s", err)
	}

	if tables.count() != 3 {
		t.Fatalf("expected to have 3 tables got %d", tables.count())
	}

	expectedTables := []table{
		{
			ID:          "order",
			Name:        "order",
			Description: "",
		}, {
			ID:          "product",
			Name:        "product",
			Description: "",
		}, {
			ID:          "order_line",
			Name:        "order_line",
			Description: "",
		},
	}

	for _, e := range expectedTables {
		exists := false
		for _, tb := range tables {
			if tb.ID == e.ID {
				exists = true
				if !reflect.DeepEqual(tb, e) {
					t.Errorf("expected table data to be %+v; got %+v instead", e, t)
				}
				break
			}
		}
		if !exists {
			t.Errorf("table %s does not exist after setup", e.Name)
		}
	}

	databaseInfo, err := storage.GetDatabaseInfo()
	if err != nil {
		t.Fatalf("we shouldn't get an error from GetDatabaseInfo; got %s", err)
	}
	if equal, _ := compareStoredDatabaseInfoWithConf(databaseInfo, conf); !equal {
		t.Errorf("database info (%+v) differs from current conf (%+v)", databaseInfo, conf)
	}

	orderTable, _ := tables.get("order")
	err = storage.UpdateAddTableDescription(orderTable.ID, "I am a cool table.")
	if err != nil {
		t.Fatalf("we shouldn't get any error from UpdateAddTableDescription; got %s", err)
	}

	tables, _ = storage.GetTables()
	updatedTable, _ := tables.get("order")
	if updatedTable.Description != "I am a cool table." {
		t.Errorf("expected tables description to be (%s); got %s instead", "I am a cool table.",
			updatedTable.Description)
	}

	columns, err := storage.GetColumns()
	if err != nil {
		t.Fatalf("we shouldn't get an error from GetColumns; got %s", err)
	}

	if len(columns) != 7 {
		t.Fatalf("expected 7 columns got %d", len(columns))
	}

	productNameCol, err := columns.getByColNameAndTableName("name", "product")
	if err != nil {
		t.Fatalf("we shouldn't get an error from getByColNameAndTableName; got %s", err)
	}

	err = storage.UpdateAddColumnDescription(productNameCol.ID, "I have a nice name.")
	if err != nil {
		t.Fatalf("we shouldn't get an error from UpdateAddColumnDescription; got %s", err)
	}

	columns, err = storage.GetColumns()
	if err != nil {
		t.Fatalf("we shouldn't get an error from GetColumns; got %s", err)
	}

	productNameCol, _ = columns.getByColNameAndTableName("name", "product")
	if productNameCol.Description != "I have a nice name." {
		t.Errorf("expected description to be (%s) in column %s in table %s got %s", "I have a nice name.",
			productNameCol.Name, productNameCol.TBName, productNameCol.Description)
	}
}
