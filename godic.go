package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var DB *sql.DB
var _logger *log.Logger

var allowedDrivers = [2]string{
	"mysql",
	"postgres",
}

var indexHtml *template.Template
var onceHtmlInit sync.Once

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
	mux.HandleFunc("/check-changes", checkDatabaseChanges(storage, conf))
	mux.HandleFunc("/sync-db", syncDatabase(storage, conf))
	mux.HandleFunc("/create-domain", CreateDomain(storage))
	mux.HandleFunc("/get-domains", GetDomains(storage))
	mux.HandleFunc("/link-table-with-domain", LinkTableWithDomain(storage))
	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.HandleFunc("/js/app.js", serveJSDevelopment())
	mux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./assets"))))
	mux.Handle("/react-compiled/", http.StripPrefix("/react-compiled", http.FileServer(http.Dir("./react"))))
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

		production := false
		if onProduction, _ := strconv.ParseBool(os.Getenv("PRODUCTION")); onProduction {
			production = onProduction
		}

		if production {
			onceHtmlInit.Do(func() {
				indexHtml = template.Must(template.ParseFiles("assets/index.html"))
			})
		} else {
			sb, err := Asset("assets/index.html")
			if err != nil {
				_logger.Println(err)
				http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
				return
			}
			indexHtml, err = template.New("").Parse(string(sb))
			if err != nil {
				_logger.Println(err)
				http.Error(w, fmt.Sprintf("Template error: %s", err), http.StatusInternalServerError)
				return
			}
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
			Production   bool
		}{
			info,
			tables,
			cols,
			production,
		}

		err = indexHtml.Execute(w, data)
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

func checkDatabaseChanges(repo Repository, conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		newTables, err := getNewTablesChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		deletedTables, err := getDeletedTablesChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		colChanges, err := getColumnChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		newCols, err := getNewColumnChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		deletedCols, err := getDeletedColumnsChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		responseData := struct {
			NewTables      []string        `json:"new_tables"`
			DeletedTables  []string        `json:"deleted_tables"`
			ColumnChanges  []columnChanges `json:"column_changes"`
			NewColumns     []newColumn     `json:"new_columns"`
			DeletedColumns []deletedColumn `json:"deleted_columns"`
		}{
			newTables,
			deletedTables,
			colChanges,
			newCols,
			deletedCols,
		}

		sb, err := json.MarshalIndent(responseData, "", strings.Repeat(" ", 3))
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(sb)
		if err != nil {
			_logger.Println(err)
		}
	}
}

func syncDatabase(repo Repository, conf *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		// Let's remove the tables that do not exist anymore.
		deletedTables, err := getDeletedTablesChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		for _, dt := range deletedTables {
			err := repo.RemoveTable(dt)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		}

		// Let's add the new tables and their columns metadata.
		newTables, err := getNewTablesChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		for _, nt := range newTables {
			t := table{Name: nt}
			err = repo.AddTable(t)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			cols, err := getTableColumns(nt, conf)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			for _, col := range cols {
				colMeta, err := columnMetadataBuilder(nt, col, conf)
				if err != nil {
					_logger.Println(err)
					http.Error(w, http.StatusText(500), http.StatusInternalServerError)
					return
				}
				err = repo.AddColMetaData(nt, colMeta)
				if err != nil {
					_logger.Println(err)
					http.Error(w, http.StatusText(500), http.StatusInternalServerError)
					return
				}
			}
		}

		// Let update the existing columns with the new changes.
		// For this, we are going to remove the old column and just create the same column, but with the updates.
		colChanges, err := getColumnChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		for _, change := range colChanges {
			storedColMetadata := change.colMetadata
			currentCol, err := getTableColumn(storedColMetadata.Name, storedColMetadata.TBName, conf)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			currentColMetadata, err := columnMetadataBuilder(storedColMetadata.TBName, currentCol, conf)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			err = repo.RemoveColMetadata(storedColMetadata.ID)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			err = repo.AddColMetaData(currentColMetadata.TBName, currentColMetadata)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		}

		// Let's add the new columns metadata of existing tables.
		newCols, err := getNewColumnChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		for _, nc := range newCols {
			col, err := getTableColumn(nc.Name, nc.Table, conf)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			colMetadata, err := columnMetadataBuilder(nc.Table, col, conf)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
			err = repo.AddColMetaData(nc.Table, colMetadata)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		}

		// Let's remove the deleted existing columns.
		deletedCols, err := getDeletedColumnsChanges(repo, conf)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}
		for _, dc := range deletedCols {
			err = repo.RemoveColMetadata(dc.ID)
			if err != nil {
				_logger.Println(err)
				http.Error(w, http.StatusText(500), http.StatusInternalServerError)
				return
			}
		}

		// If all goes well, we have successfully synced the database with the new changes.
		w.WriteHeader(http.StatusOK)
	}
}

func serveJSDevelopment() http.HandlerFunc {
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

func CreateDomain(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		var domain Domain

		err = json.Unmarshal(body, &domain)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// Let's do some cleanup.
		domain.Name = strings.TrimSpace(domain.Name)
		domain.Description = strings.TrimSpace(domain.Description)

		// Let's validate that there are no missing fields.
		if domain.Name == "" || domain.Description == "" {
			http.Error(w, fmt.Sprintf("Missing fields: %+v", domain), http.StatusBadRequest)
			return
		}

		// Let's try to save the new domain in the db.
		err = repo.CreateDomain(domain)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		// If all goes well. We reply with an OK status.
		w.WriteHeader(http.StatusOK)
	}
}

func GetDomains(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		domains, err := repo.GetDomains()
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		sb, err := json.Marshal(domains)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(sb)
		if err != nil {
			_logger.Println(err)
		}
	}
}

func LinkTableWithDomain(repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
			return
		}

		requestData := struct {
			TableID    string `json:"table_id"`
			DomainName string `json:"domain_name"`
		}{}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		err = json.Unmarshal(body, &requestData)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		requestData.TableID = strings.TrimSpace(requestData.TableID)
		requestData.DomainName = strings.TrimSpace(requestData.DomainName)

		if requestData.TableID == "" || requestData.DomainName == "" {
			http.Error(w, fmt.Sprintf("Missing fields: %+v", requestData), http.StatusBadRequest)
			return
		}

		err = repo.LinkTableWithDomain(requestData.TableID, requestData.DomainName)
		if err != nil {
			_logger.Println(err)
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
