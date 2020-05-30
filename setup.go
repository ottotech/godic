package main

import (
	"github.com/jimsmart/schema"
	"strings"
)

// addDatabaseMetaData stores in repository all the database metadata.
func addDatabaseMetaData(storage Repository) error {
	if exists, err := storage.IsDatabaseMetaDataAdded(*dbName); err != nil {
		return err
	} else if !exists {
		data := dbInfo{
			Name:     *dbName,
			User:     *dbUser,
			Host:     *dbHost,
			Port:     *dbPort,
			Password: *dbPassword,
			Driver:   *dbDriver,
		}
		err := storage.AddDatabaseInfo(data)
		if err != nil {
			return err
		}

		tableNames, err := schema.TableNames(DB)
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
			tableColumns, err := schema.Table(DB, tableNames[i])
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
					colMeta.ENUMValues = strings.Split(enum.EnumValue, ",")
				}

				if exists := uniques.exists(colMeta.Name, tableNames[i]); exists {
					uCol, err := uniques.get(colMeta.Name, tableNames[i])
					if err != nil {
						return err
					}
					colMeta.IsUnique = true
					colMeta.UniqueIndexDefinition = uCol.UniqueDefinition
				}

				t := table{Name: tableNames[i]}
				err = storage.AddTable(t)
				if err != nil {
					return err
				}

				err = storage.AddColMetaData(tableNames[i], colMeta)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
