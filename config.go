package main

import (
	"bytes"
	"flag"
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

	return true, ""
}

// ParseFlags parses the flags given to the godic app.
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
