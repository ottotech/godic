package main

import (
	"bytes"
	"flag"
)

// Config holds the different configuration options of the database as well as some option for the godic app.
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
}

// ParseFlags parses the
func ParseFlags(programName string, args []string) (config *Config, output string, err error) {
	flags := flag.NewFlagSet(programName, flag.ContinueOnError)
	var buf bytes.Buffer
	flags.SetOutput(&buf)

	var conf Config
	flags.IntVar(&conf.ServerPort, "server_port", 8080, "port used for http server")
	flags.StringVar(&conf.DatabaseUser, "db_user", "", "database user")
	flags.StringVar(&conf.DatabasePassword, "db_password", "", "database password")
	flags.StringVar(&conf.DatabaseHost, "db_host", "", "database host")
	flags.IntVar(&conf.DatabasePort, "db_port", 5432, "database port")
	flags.StringVar(&conf.DatabaseName, "db_name", "", "database name")
	flags.StringVar(&conf.DatabaseDriver, "db_driver", "", "database driver")
	flags.StringVar(&conf.DatabaseSchema, "db_schema", "public", "database schema")
	flags.BoolVar(&conf.ForceDelete, "force_delete", false, "deletes completely any stored metadata of a database in order to start fresh")

	err = flags.Parse(args)
	if err != nil {
		return nil, buf.String(), err
	}

	return &conf, buf.String(), nil
}
