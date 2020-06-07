package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
)

var (
	mysqlTestDb *sql.DB
	psqlTestDb  *sql.DB
)

var (
	psqlDatabaseUri  = "user=test password=secret host=localhost port=5555 dbname=%s sslmode=disable"
	mysqlDatabaseUri = "root:secret@tcp(localhost:3333)/%s?multiStatements=true"
)

const (
	// psql database constant vars
	testPsqlDatabaseName     = "test_db"
	testPsqlDatabasePassword = "secret"
	testPsqlDatabasePort     = 5555
	testPsqlDatabaseUser     = "test"
	testPsqlDatabaseHost     = "localhost"
	testPsqlDatabaseDriver   = "postgres"
	testPsqlDatabaseSchema   = "public"

	// mysql database constant vars
	testMysqlDatabaseName     = "test_db"
	testMysqlDatabasePassword = "secret"
	testMysqlDatabasePort     = 3333
	testMysqlDatabaseUser     = "test"
	testMysqlDatabaseHost     = "localhost"
	testMysqlDatabaseDriver   = "mysql"
	testMysqlDatabaseSchema   = "test_db"
)

func createPsqlConf() *Config {
	return &Config{
		ServerPort:       0000,
		DatabaseUser:     testPsqlDatabaseUser,
		DatabasePassword: testPsqlDatabasePassword,
		DatabaseHost:     testPsqlDatabaseHost,
		DatabasePort:     testPsqlDatabasePort,
		DatabaseName:     testPsqlDatabaseName,
		DatabaseDriver:   testPsqlDatabaseDriver,
		DatabaseSchema:   testPsqlDatabaseSchema,
		ForceDelete:      false,
	}
}

func createPsqlDatabase() error {
	db, err := sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, "postgres"))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", testPsqlDatabaseName))
	if err != nil {
		return err
	}
	db.Close()

	db, err = sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, testPsqlDatabaseName))
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

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s;", testPsqlDatabaseName))
	if err != nil {
		return err
	}
	return nil
}

func createMysqlConf() *Config {
	return &Config{
		ServerPort:       0000,
		DatabaseUser:     testMysqlDatabaseUser,
		DatabasePassword: testMysqlDatabasePassword,
		DatabaseHost:     testMysqlDatabaseHost,
		DatabasePort:     testMysqlDatabasePort,
		DatabaseName:     testMysqlDatabaseName,
		DatabaseDriver:   testMysqlDatabaseDriver,
		DatabaseSchema:   testMysqlDatabaseSchema,
		ForceDelete:      false,
	}
}

func createMysqlDatabase() error {
	db, err := sql.Open("mysql", fmt.Sprintf(mysqlDatabaseUri, "mysql"))
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s;", testMysqlDatabaseName))
	if err != nil {
		return err
	}
	db.Close()

	db, err = sql.Open("mysql", fmt.Sprintf(mysqlDatabaseUri, testMysqlDatabaseName))
	if err != nil {
		return err
	}
	defer db.Close()

	q1 := `
		CREATE TABLE` + " `order` " + ` 
		  ( 
			 id BIGINT UNSIGNED AUTO_INCREMENT,
			 CONSTRAINT id
			      unique (id)
		  );

		ALTER TABLE` + " `order` " +
		`ADD PRIMARY KEY (id);
	`
	q2 := `
		CREATE TABLE product 
		  ( 
			 id   BIGINT UNSIGNED AUTO_INCREMENT,
			 name VARCHAR(200) NOT NULL,
			 counting_option ENUM('unit', 'decimal') NOT NULL,
			 CONSTRAINT product_id_uindex
					unique (id),
			 CONSTRAINT product_name_uindex
					unique (name)
		  );

		ALTER TABLE product ADD PRIMARY KEY (id);
	`
	q3 := `
		CREATE TABLE order_line 
		  ( 
			 id         BIGINT UNSIGNED AUTO_INCREMENT, 
			 order_id   BIGINT UNSIGNED, 
			 product_id BIGINT UNSIGNED, 
			 CONSTRAINT order_line_id_uindex
					unique (id),
             CONSTRAINT order_line_order_id_fk
                 FOREIGN KEY (order_id) REFERENCES` + " `order` " + `(id)
                     ON DELETE RESTRICT,
             CONSTRAINT order_line_product_id_fk
                 FOREIGN KEY (product_id) REFERENCES product (id)
					 ON DELETE RESTRICT
		  ); 

		ALTER TABLE` + " `order_line` " +
		`ADD PRIMARY KEY (id);
	`

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}

	for _, q := range []string{q1, q2, q3} {
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

func removeMysqlDatabase() error {
	db, err := sql.Open("mysql", fmt.Sprintf(mysqlDatabaseUri, "mysql"))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE %s;", testPsqlDatabaseName))
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	var err error
	err = createPsqlDatabase()
	if err != nil {
		removePsqlDatabase()
		log.Fatalln(err)
	}

	err = createMysqlDatabase()
	if err != nil {
		removeMysqlDatabase()
		log.Fatalln(err)
	}

	psqlDB, err := sql.Open("postgres", fmt.Sprintf(psqlDatabaseUri, testPsqlDatabaseName))
	if err != nil {
		log.Fatalln(err)
	}
	psqlTestDb = psqlDB

	mysqlDB, err := sql.Open("mysql", fmt.Sprintf(mysqlDatabaseUri, testMysqlDatabaseName))
	if err != nil {
		log.Fatalln(err)
	}
	mysqlTestDb = mysqlDB

	code := m.Run()

	psqlTestDb.Close()
	mysqlTestDb.Close()

	err = removePsqlDatabase()
	if err != nil {
		log.Println(err)
	}

	err = removeMysqlDatabase()
	if err != nil {
		log.Println(err)
	}

	os.Exit(code)
}
