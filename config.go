package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ian-kent/envconf"
)

// Config holds the different configuration options of the database as well as some options for the godic app.
type Config struct {
	ServerPort       int    `json:"server_port"`
	DatabaseUser     string `json:"database_user"`
	DatabasePassword string `json:"database_password"`
	DatabaseHost     string `json:"database_host"`
	DatabasePort     int    `json:"database_port"`
	DatabaseName     string `json:"database_name"`
	DatabaseDriver   string `json:"database_driver"`
	DatabaseSchema   string `json:"database_schema"`
	ForceDelete      bool   `json:"force_delete"`
	Storage          string `json:"storage"`
	MongoDB          string `json:"mongo_db"`
	MongoUri         string `json:"mongo_uri"`
}

// validate validates the configuration options given to Config.
func (c *Config) validate() (ok bool, msg string) {
	msg = "There are some options missing from the flags given to run godic, please refer to -h to check " +
		"godic flags usage."

	if c.ServerPort == 0 {
		return
	}

	if c.DatabaseUser == "" || c.DatabasePassword == "" || c.DatabaseHost == "" || c.DatabasePort == 0 ||
		c.DatabaseSchema == "" || c.DatabaseDriver == "" || c.DatabaseName == "" {
		return
	}

	if c.Storage != "" && c.Storage != "mongo" {
		return false, fmt.Sprintf("The only available storage option is %q.", "mongo")
	}

	if c.Storage == "mongo" && c.MongoDB == "" {
		return false, "You need to define the name you want to use for your mongo database."
	}

	if c.Storage == "mongo" && c.MongoUri == "" {
		return false, "You need to define the uri to connect to your mongo db instance."
	}

	return true, ""
}

// ParseFlags parses the flags given to the godic app.
func ParseFlags(programName string, args []string) (config *Config, output string, err error) {
	flags := flag.NewFlagSet(programName, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var conf Config
	flags.IntVar(&conf.ServerPort, "server_port", envconf.FromEnvP("GODIC_SERVER_PORT", 8080).(int), "port used for http server")
	flags.StringVar(&conf.DatabaseUser, "db_user", envconf.FromEnvP("GODIC_DB_USER", "").(string), "database user")
	flags.StringVar(&conf.DatabasePassword, "db_password", envconf.FromEnvP("GODIC_DB_PASSWORD", "").(string), "database password")
	flags.StringVar(&conf.DatabaseHost, "db_host", envconf.FromEnvP("GODIC_DB_HOST", "").(string), "database host")
	flags.IntVar(&conf.DatabasePort, "db_port", envconf.FromEnvP("GODIC_DB_PORT", 5432).(int), "database port")
	flags.StringVar(&conf.DatabaseName, "db_name", envconf.FromEnvP("GODIC_DB_NAME", "").(string), "database name")
	flags.StringVar(&conf.DatabaseDriver, "db_driver", envconf.FromEnvP("GODIC_DB_DRIVER", "").(string), "database driver")
	flags.StringVar(&conf.DatabaseSchema, "db_schema", envconf.FromEnvP("GODIC_DB_SCHEMA", "public").(string), "database schema")
	flags.BoolVar(&conf.ForceDelete, "force_delete", envconf.FromEnvP("GODIC_FORCE_DELETE", false).(bool), "deletes completely any stored metadata of a database in order to start fresh")
	flags.StringVar(&conf.Storage, "storage", envconf.FromEnvP("GODIC_DB_STORAGE", "json").(string), "storage type for saving data")
	flags.StringVar(&conf.MongoDB, "mongo-db", envconf.FromEnvP("GODIC_MONGO_DB", "godicdb").(string), "mongo db name where we are going to save the data")
	flags.StringVar(&conf.MongoUri, "mongo-uri", envconf.FromEnvP("GODIC_MONGO_URI", "").(string), "mongo uri fo your db instance")

	err = flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}
	return &conf, buf.String(), nil
}
