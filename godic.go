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
	"strconv"
)

var DB *sql.DB
var _logger *log.Logger

var allowedDrivers = [2]string{
	"mysql",
	"postgres",
}

const psqlDbSource string = "user=%s password=%s host=%s port=%d dbname=%s sslmode=disable"
const mysqlDbSource string = "{user}:{password}@tcp({host}:{port})/{database}"

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

	conf, output, err := ParseFlags(os.Args[0], os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Println(output)
		os.Exit(2)
	} else if err != nil {
		fmt.Println("output:\n", output)
		os.Exit(1)
	}

	if ok, msg := conf.validate(); !ok {
		log.Fatalln(msg)
	}

	err = run(conf)
	if err != nil {
		log.Fatalln(err)
	}
}

func run(conf *Config) error {
	if err := validateSqlDriver(conf); err != nil {
		return err
	}

	source := ""
	if conf.DatabaseDriver == "mysql" {
		source = formatMysqlSource(conf)
	} else if conf.DatabaseDriver == "postgres" {
		source = fmt.Sprintf(psqlDbSource, conf.DatabaseUser, conf.DatabasePassword, conf.DatabaseHost,
			conf.DatabasePort, conf.DatabaseName)
	}

	storage, err := NewJsonStorage()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(conf.DatabaseDriver, source)
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	DB = db
	log.Println("You connected to your database: ", conf.DatabaseName)

	if conf.DatabaseDriver == "postgres" {
		_, err = DB.Exec(fmt.Sprintf("SET search_path=%s", conf.DatabaseSchema))
		if err != nil {
			panic(err)
		}
	}

	err = setupInitialMetadata(storage, conf)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", index(storage))
	mux.HandleFunc("/update", updateTableDictionary(storage))
	mux.HandleFunc("/js/app.js", serveJS())
	srv := http.Server{
		Addr:    ":" + strconv.Itoa(conf.ServerPort),
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		_logger.Println(err)
		return err
	}

	return nil
}

func index(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sb, err := Asset("assets/index.html")
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

		info, err := repo.GetDatabaseInfo()
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		tables, err := repo.GetTables()
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		cols, err := repo.GetColumns()
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		data := struct {
			DatabaseInfo databaseInfo
			Tables       Tables
			Columns      ColumnsMetadata
		}{
			info,
			tables,
			cols,
		}

		err = tpl.Execute(w, data)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		}
	}
}

func updateTableDictionary(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		requestData := struct {
			TableID          string `json:"table_id"`
			TableDescription string `json:"table_description"`
			ColumnsData      []struct {
				ColID       string `json:"col_id"`
				Description string `json:"description"`
			} `json:"columns_data"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		// error managed like 500 for simplicity.
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		err = repo.UpdateAddTableDescription(requestData.TableID, requestData.TableDescription)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		for _, col := range requestData.ColumnsData {
			err = repo.UpdateAddColumnDescription(col.ColID, col.Description)
			if err != nil {
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func serveJS() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sb, err := Asset("assets/app.js")
		if err != nil {
			_logger.Println(err)
			http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
			return
		}

		_, err = w.Write(sb)
		if err != nil {
			_logger.Println(err)
		}
	}
}
