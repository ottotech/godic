package main

import (
	"fmt"
	"strings"
)

func runSetup(storage Repository) error {
	var err error

	metaDataExists, err := storage.IsDatabaseMetaDataAdded(*dbName)
	if err != nil {
		return err
	}

	if *forceDelete {
		goto DoForceDelete
	}

	if metaDataExists {
		databaseInfo, err := storage.GetDatabaseInfo()
		if err != nil {
			return err
		}
		equal, msg := compareStoredDatabaseInfoWithFlags(databaseInfo)
		if !equal {
			return fmt.Errorf("The flags provided do not match the ones stored in your "+
				"database info.\nIf you want to remove completely the data dictionary of your "+
				"previous database and start fresh you can run godic with the flag -force_delete=true "+
				"(see documentation of this flag).\n"+
				"Here some of the differences we found:\n%s", msg)
		}
		goto DoNothing
	}

	if !metaDataExists {
		databaseInfo, err := storage.GetDatabaseInfo()
		if err != nil {
			if err == ErrNoDatabaseMetaDataStored {
				goto DoSetup
			}
			return fmt.Errorf("got error when trying to run storage.GetDatabaseInfo(); %s", err)
		}
		return fmt.Errorf("there is already some metadata stored for the database %s "+
			"with the following info:\n"+
			"Port: %d\n"+
			"User: %s\n"+
			"Schema: %s\n"+
			"Host: %s\n"+
			"Password: %s\n"+
			"Driver: %s\n\n"+
			"If you want to remove this all database meta data you can run godic with the "+
			"flag -force_delete=true (see documentation for this flag)\n",
			databaseInfo.Name,
			databaseInfo.Port,
			databaseInfo.User,
			databaseInfo.Schema,
			databaseInfo.Host,
			databaseInfo.Password,
			databaseInfo.Driver,
		)
	}
DoForceDelete:
	err = storage.RemoveEverything()
	if err != nil {
		return err
	}
	goto DoSetup
DoSetup:
	err = databaseMetaDataSetup(storage)
	if err != nil {
		return err
	}
	return nil
DoNothing:
	return nil
}

// databaseMetaDataSetup stores in repository all the database metadata.
func databaseMetaDataSetup(storage Repository) error {
	dbInfo := databaseInfo{
		Name:     *dbName,
		User:     *dbUser,
		Host:     *dbHost,
		Port:     *dbPort,
		Password: *dbPassword,
		Driver:   *dbDriver,
		Schema:   *dbSchema,
	}
	err := storage.AddDatabaseInfo(dbInfo)
	if err != nil {
		return err
	}
	tableNames, err := getTableNames()
	if err != nil {
		return err
	}

	primaryKeys, err := getPrimaryKeys()
	if err != nil {
		return err
	}

	foreignKeys, err := getForeignKeys()
	if err != nil {
		return err
	}

	enums, err := getColsAndEnums()
	if err != nil {
		return err
	}

	uniques, err := getUniqueCols()
	if err != nil {
		return err
	}

	for i := range tableNames {
		t := table{Name: tableNames[i]}
		err = storage.AddTable(t)
		if err != nil {
			return err
		}

		tableColumns, err := getTableColumns(tableNames[i])
		if err != nil {
			return err
		}

		for _, col := range tableColumns {
			colMeta := colMetaData{}
			colMeta.Name = col.Name()
			colMeta.DBType = col.DatabaseTypeName()
			colMeta.Nullable = parseNullableFromCol(col)
			colMeta.GoType = col.ScanType().String()
			colMeta.Length = parseLengthFromCol(col)
			colMeta.TBName = tableNames[i]

			if isPK := primaryKeys.exists(colMeta.Name, tableNames[i]); isPK {
				colMeta.IsPrimaryKey = true
			}

			if isFK := foreignKeys.exists(colMeta.Name, tableNames[i]); isFK {
				fk, err := foreignKeys.get(colMeta.Name, tableNames[i])
				if err != nil {
					return err
				}
				colMeta.IsForeignKey = true
				colMeta.TargetTableFK = fk.TargetTable
				colMeta.DeleteRule = fk.DeleteRule
				colMeta.UpdateRule = fk.UpdateRule
			}

			if hasEnum := enums.exists(colMeta.Name, tableNames[i]); hasEnum {
				enum, err := enums.get(colMeta.Name, tableNames[i])
				if err != nil {
					return err
				}
				colMeta.HasENUM = true
				colMeta.ENUMName = enum.EnumName
				colMeta.ENUMValues = strings.Split(enum.EnumValues, ",")
			}

			if hasUniqueIndex := uniques.exists(colMeta.Name, tableNames[i]); hasUniqueIndex {
				uCol, err := uniques.get(colMeta.Name, tableNames[i])
				if err != nil {
					return err
				}
				colMeta.IsUnique = true
				colMeta.UniqueIndexDefinition = uCol.UniqueDefinition
			}

			err = storage.AddColMetaData(tableNames[i], colMeta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
