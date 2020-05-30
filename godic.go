package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
)

var DB *sql.DB
var _logger *log.Logger

const dbSource string = "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable"

var (
	port       = flag.String("server_port", "8080", "port used for http server")
	dbUser     = flag.String("db_user", "", "database user")
	dbPassword = flag.String("db_password", "", "database password")
	dbHost     = flag.String("db_host", "", "database host")
	dbPort     = flag.Int("db_port", 5432, "database port")
	dbName     = flag.String("db_name", "", "database name")
	dbDriver   = flag.String("db_driver", "postgres", "database driver")
	dbSchema   = flag.String("db_schema", "public", "database schema for search_path")
)

func main() {
	logFile, err := os.OpenFile("./error.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	_logger = log.New(logFile, "Error Logger:\t", log.Ldate|log.Ltime|log.Lshortfile)
	defer func() {
		err := logFile.Close()
		if err != nil {
			_logger.Println(err)
		}
	}()

	flag.Parse()
	source := fmt.Sprintf(dbSource, *dbUser, *dbPassword, *dbHost, *dbPort, *dbName)
	storage, err := NewJsonStorage()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(*dbDriver, source)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	_, err = db.Exec(fmt.Sprintf("SET search_path=%s", *dbSchema))
	if err != nil {
		panic(err)
	}

	DB = db
	_logger.Println("You connected to your database: ", *dbName)
	err = addDatabaseMetaData(storage)
	if err != nil {
		_logger.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index())
	srv := http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		_logger.Println(err)
	}
}

func index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sb, err := Asset("index.html")
		if err != nil {
			_logger.Println(err)
			http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
			return
		}
		tpl, err := template.New("").Parse(string(sb))
		if err != nil {
			_logger.Println(err)
			http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
			return
		}
		err = tpl.Execute(w, nil)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		}
	}
}
