package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/net/context"
	"log"
	"os"
	"reflect"
	"testing"
)

var psqlTestDb *sql.DB

const (
	testDatabaseName     = "test_db"
	testDatabasePassword = "secret"
	testDatabasePort     = 5555
	testDatabaseUser     = "test"
	testDatabaseHost     = "localhost"
	testDatabaseDriver   = "postgres"
	testDatabaseSchema   = "public"
)

var psqlDatabaseUri = "user=test password=secret host=localhost port=5555 dbname=%s sslmode=disable"

func createConf() *Config {
	return &Config{
		ServerPort:       0000,
		DatabaseUser:     testDatabaseUser,
		DatabasePassword: testDatabasePassword,
		DatabaseHost:     testDatabaseHost,
		DatabasePort:     testDatabasePort,
		DatabaseName:     testDatabaseName,
		DatabaseDriver:   testDatabaseDriver,
		DatabaseSchema:   testDatabaseSchema,
		ForceDelete:      false,
	}
}

func createPsqlDatabase() error {
	db, err := sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, "postgres"))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", testDatabaseName))
	if err != nil {
		return err
	}
	db.Close()

	db, err = sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, testDatabaseName))
	if err != nil {
		return err
	}
	defer db.Close()

	q1 := `
		CREATE TABLE "order" 
		  ( 
			 id SERIAL NOT NULL CONSTRAINT order_pk PRIMARY KEY
		  ); 
		
		CREATE UNIQUE INDEX order_id_uindex 
		  ON "order" (id); 
	`
	q2 := `
		CREATE TYPE counting_option AS enum ('unit', 'decimal'); 
	`
	q3 := `
		CREATE TABLE product 
		  ( 
			 id   SERIAL NOT NULL CONSTRAINT product_pk PRIMARY KEY, 
			 name VARCHAR(200) NOT NULL,
             counting_option counting_option NOT NULL
		  ); 
	
		CREATE UNIQUE INDEX product_id_uindex ON product (id);
	
		CREATE UNIQUE INDEX product_name_uindex ON product (name);
	`
	q4 := `
		CREATE TABLE order_line 
		  ( 
			 id         SERIAL NOT NULL CONSTRAINT order_line_pk PRIMARY KEY, 
			 order_id   INTEGER NOT NULL CONSTRAINT order_line_order_id_fk REFERENCES 
			 "order" 
			 ON DELETE RESTRICT, 
			 product_id INTEGER NOT NULL CONSTRAINT order_line_product_id_fk REFERENCES 
			 product ON DELETE 
			 RESTRICT 
		  ); 
	
		CREATE UNIQUE INDEX order_line_id_uindex ON order_line (id);
	`

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	for _, q := range []string{q1, q2, q3, q4} {
		_, err = tx.Exec(q)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return err
			}
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func removePsqlDatabase() error {
	db, err := sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, "postgres"))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s;", testDatabaseName))
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	err := createPsqlDatabase()
	if err != nil {
		log.Fatalln(err)
	}

	db, err := sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, testDatabaseName))
	if err != nil {
		log.Fatalln(err)
	}
	psqlTestDb = db

	code := m.Run()

	psqlTestDb.Close()

	err = removePsqlDatabase()
	if err != nil {
		log.Println(err)
	}

	os.Exit(code)
}

func Test_getTableNames_helper_func(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

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

func Test_getTableColumns_helpers_func(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

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

func Test_getPrimaryKeys(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

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

func Test_getForeignKeys_helper_func(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

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

func Test_getColsAndEnums_helper_func(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

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
		EnumName:   "counting_option",
		EnumValues: "unit, decimal",
	}

	if !reflect.DeepEqual(expectedEnum, enums[0]) {
		t.Errorf("expected enum %+v; got %+v", expectedEnum, enums[0])
	}
}

func Test_getUniqueCols_helper_func(t *testing.T) {
	originalDB := DB
	DB = psqlTestDb
	defer func(original *sql.DB) {
		DB = original
	}(originalDB)

	conf := createConf()

	uniqueColumns, err := getUniqueCols(conf)
	if err != nil {
		t.Fatalf("we shouldn't get an error when calling getUniqueCols; got %s", err)
	}

	if len(uniqueColumns) != 4 {
		t.Errorf("we expected 4 unique columns got %d", len(uniqueColumns))
	}
}
