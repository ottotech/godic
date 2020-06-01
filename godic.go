package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"html/template"
	"log"
	"net/http"
	"os"
)

var DB *sql.DB
var _logger *log.Logger

var allowedDrivers = [2]string{
	"mysql",
	"postgres",
}

const psqlDbSource string = "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable"
const mysqlDbSource string = "{user}:{password}@tcp({host}:{port})/{database}"

var (
	port        = flag.String("server_port", "8080", "port used for http server")
	dbUser      = flag.String("db_user", "", "database user")
	dbPassword  = flag.String("db_password", "", "database password")
	dbHost      = flag.String("db_host", "", "database host")
	dbPort      = flag.Int("db_port", 5432, "database port")
	dbName      = flag.String("db_name", "", "database name")
	dbDriver    = flag.String("db_driver", "postgres", "database driver")
	dbSchema    = flag.String("db_schema", "public", "database schema for search_path")
	forceDelete = flag.Bool("force_delete", false, "deletes completely any stored metadata of a database in order to start fresh")
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

	source := ""
	if *dbDriver == "mysql" {
		source = formatMysqlSource()
	} else if *dbDriver == "postgres" {
		source = fmt.Sprintf(psqlDbSource, *dbUser, *dbPassword, *dbHost, *dbPort, *dbName)
	}

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

	DB = db
	log.Println("You connected to your database: ", *dbName)

	if *dbDriver == "postgres" {
		_, err = DB.Exec(fmt.Sprintf("SET search_path=%s", *dbSchema))
		if err != nil {
			panic(err)
		}
	}

	err = runSetup(storage)
	if err != nil {
		log.Fatalln(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index())
	mux.HandleFunc("/update-add-table-description", updateAddTableDescriptionHandler(storage))
	mux.HandleFunc("/update-add-column-description", updateAddColumnDescriptionHandler(storage))
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

func updateAddTableDescriptionHandler(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		requestData := struct {
			TableID     string `json:"table_id"`
			Description string `json:"description"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			// error managed like 500 for simplicity.
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		err = repo.UpdateAddTableDescription(requestData.TableID, requestData.Description)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func updateAddColumnDescriptionHandler(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		requestData := struct {
			ColumnID    string `json:"column_id"`
			Description string `json:"description"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			// error managed like 500 for simplicity.
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		err = repo.UpdateAddColumnDescription(requestData.ColumnID, requestData.Description)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}
